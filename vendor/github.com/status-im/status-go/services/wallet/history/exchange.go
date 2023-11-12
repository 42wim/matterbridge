package history

import (
	"errors"
	"sync"
	"time"

	"github.com/status-im/status-go/services/wallet/market"
)

type tokenType = string
type currencyType = string
type yearType = int

type allTimeEntry struct {
	value          float32
	startTimestamp int64
	endTimestamp   int64
}

// Exchange caches conversion rates in memory on a daily basis
type Exchange struct {
	// year map keeps a list of values with days as index in the slice for the corresponding year (key) starting from the first to the last available
	cache map[tokenType]map[currencyType]map[yearType][]float32
	// special case for all time information
	allTimeCache map[tokenType]map[currencyType][]allTimeEntry
	fetchMutex   sync.Mutex

	marketManager *market.Manager
}

func NewExchange(marketManager *market.Manager) *Exchange {
	return &Exchange{
		cache:         make(map[tokenType]map[currencyType]map[yearType][]float32),
		marketManager: marketManager,
	}
}

// GetExchangeRate returns the exchange rate from token to currency in the day of the given date
// if none exists returns "missing <element>" error
func (e *Exchange) GetExchangeRateForDay(token tokenType, currency currencyType, date time.Time) (float32, error) {
	e.fetchMutex.Lock()
	defer e.fetchMutex.Unlock()

	currencyMap, found := e.cache[token]
	if !found {
		return 0, errors.New("missing token")
	}

	yearsMap, found := currencyMap[currency]
	if !found {
		return 0, errors.New("missing currency")
	}

	year := date.Year()
	valueForDays, found := yearsMap[year]
	if !found {
		// Search closest in all time
		allCurrencyMap, found := e.allTimeCache[token]
		if !found {
			return 0, errors.New("missing token in all time data")
		}

		allYearsMap, found := allCurrencyMap[currency]
		if !found {
			return 0, errors.New("missing currency in all time data")
		}
		for _, entry := range allYearsMap {
			if entry.startTimestamp <= date.Unix() && entry.endTimestamp > date.Unix() {
				return entry.value, nil
			}
		}
		return 0, errors.New("missing entry")
	}

	day := date.YearDay()
	if day >= len(valueForDays) {
		return 0, errors.New("missing day")
	}
	return valueForDays[day], nil
}

// fetchAndCacheRates fetches and in memory cache exchange rates for this and last year
func (e *Exchange) FetchAndCacheMissingRates(token tokenType, currency currencyType) error {
	// Protect REST calls also to prevent fetching the same token/currency twice
	e.fetchMutex.Lock()
	defer e.fetchMutex.Unlock()

	// Allocate missing values
	currencyMap, found := e.cache[token]
	if !found {
		currencyMap = make(map[currencyType]map[yearType][]float32)
		e.cache[token] = currencyMap
	}

	yearsMap, found := currencyMap[currency]
	if !found {
		yearsMap = make(map[yearType][]float32)
		currencyMap[currency] = yearsMap
	}

	currentTime := time.Now().UTC()
	endOfPrevYearTime := time.Date(currentTime.Year()-1, 12, 31, 23, 0, 0, 0, time.UTC)

	daysToFetch := extendDaysSliceForYear(yearsMap, endOfPrevYearTime)

	curYearTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
	daysToFetch += extendDaysSliceForYear(yearsMap, curYearTime)
	if daysToFetch == 0 {
		return nil
	}

	res, err := e.marketManager.FetchHistoricalDailyPrices(token, currency, daysToFetch, false, 1)
	if err != nil {
		return err
	}

	for i := 0; i < len(res); i++ {
		t := time.Unix(res[i].Timestamp, 0).UTC()
		yearDayIndex := t.YearDay() - 1
		yearValues, found := yearsMap[t.Year()]
		if found && yearDayIndex < len(yearValues) {
			yearValues[yearDayIndex] = float32(res[i].Value)
		}
	}

	// Fetch all time
	allTime, err := e.marketManager.FetchHistoricalDailyPrices(token, currency, 1, true, 30)
	if err != nil {
		return err
	}

	if e.allTimeCache == nil {
		e.allTimeCache = make(map[tokenType]map[currencyType][]allTimeEntry)
	}
	_, found = e.allTimeCache[token]
	if !found {
		e.allTimeCache[token] = make(map[currencyType][]allTimeEntry)
	}

	// No benefit to fetch intermendiate values, overwrite historical
	e.allTimeCache[token][currency] = make([]allTimeEntry, 0)

	for i := 0; i < len(allTime) && allTime[i].Timestamp < res[0].Timestamp; i++ {
		if allTime[i].Value > 0 {
			var endTimestamp int64
			if i+1 < len(allTime) {
				endTimestamp = allTime[i+1].Timestamp
			} else {
				endTimestamp = res[0].Timestamp
			}
			e.allTimeCache[token][currency] = append(e.allTimeCache[token][currency],
				allTimeEntry{
					value:          float32(allTime[i].Value),
					startTimestamp: allTime[i].Timestamp,
					endTimestamp:   endTimestamp,
				})
		}
	}

	return nil
}

func extendDaysSliceForYear(yearsMap map[yearType][]float32, untilTime time.Time) (daysToFetch int) {
	year := untilTime.Year()
	_, found := yearsMap[year]
	if !found {
		yearsMap[year] = make([]float32, untilTime.YearDay())
		return untilTime.YearDay()
	}

	// Just extend the slice if needed
	missingDays := untilTime.YearDay() - len(yearsMap[year])
	yearsMap[year] = append(yearsMap[year], make([]float32, missingDays)...)
	return missingDays
}
