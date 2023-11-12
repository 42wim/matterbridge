package market

import (
	"sync"
	"time"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

const (
	EventMarketStatusChanged walletevent.EventType = "wallet-market-status-changed"
)

type DataPoint struct {
	Price     float64
	UpdatedAt int64
}

type DataPerTokenAndCurrency = map[string]map[string]DataPoint

type Manager struct {
	main            thirdparty.MarketDataProvider
	fallback        thirdparty.MarketDataProvider
	feed            *event.Feed
	priceCache      DataPerTokenAndCurrency
	priceCacheLock  sync.RWMutex
	IsConnected     bool
	LastCheckedAt   int64
	IsConnectedLock sync.RWMutex
}

func NewManager(main thirdparty.MarketDataProvider, fallback thirdparty.MarketDataProvider, feed *event.Feed) *Manager {
	hystrix.ConfigureCommand("marketClient", hystrix.CommandConfig{
		Timeout:               10000,
		MaxConcurrentRequests: 100,
		SleepWindow:           300000,
		ErrorPercentThreshold: 25,
	})

	return &Manager{
		main:          main,
		fallback:      fallback,
		feed:          feed,
		priceCache:    make(DataPerTokenAndCurrency),
		IsConnected:   true,
		LastCheckedAt: time.Now().Unix(),
	}
}

func (pm *Manager) setIsConnected(value bool) {
	pm.IsConnectedLock.Lock()
	defer pm.IsConnectedLock.Unlock()
	pm.LastCheckedAt = time.Now().Unix()
	if value != pm.IsConnected {
		message := "down"
		if value {
			message = "up"
		}
		pm.feed.Send(walletevent.Event{
			Type:     EventMarketStatusChanged,
			Accounts: []common.Address{},
			Message:  message,
			At:       time.Now().Unix(),
		})
	}
	pm.IsConnected = value
}

func (pm *Manager) makeCall(main func() (any, error), fallback func() (any, error)) (any, error) {
	resultChan := make(chan any, 1)
	errChan := hystrix.Go("marketClient", func() error {
		res, err := main()
		if err != nil {
			return err
		}
		pm.setIsConnected(true)
		resultChan <- res
		return nil
	}, func(err error) error {
		if pm.fallback == nil {
			return err
		}

		res, err := fallback()
		if err != nil {
			pm.setIsConnected(false)
			return err
		}
		pm.setIsConnected(true)
		resultChan <- res
		return nil
	})
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:

		return nil, err
	}
}

func (pm *Manager) FetchHistoricalDailyPrices(symbol string, currency string, limit int, allData bool, aggregate int) ([]thirdparty.HistoricalPrice, error) {
	prices, err := pm.makeCall(
		func() (any, error) {
			return pm.main.FetchHistoricalDailyPrices(symbol, currency, limit, allData, aggregate)
		},
		func() (any, error) {
			return pm.fallback.FetchHistoricalDailyPrices(symbol, currency, limit, allData, aggregate)
		},
	)
	if err != nil {
		return nil, err
	}

	return prices.([]thirdparty.HistoricalPrice), nil
}

func (pm *Manager) FetchHistoricalHourlyPrices(symbol string, currency string, limit int, aggregate int) ([]thirdparty.HistoricalPrice, error) {
	prices, err := pm.makeCall(
		func() (any, error) {
			return pm.main.FetchHistoricalHourlyPrices(symbol, currency, limit, aggregate)
		},
		func() (any, error) {
			return pm.fallback.FetchHistoricalHourlyPrices(symbol, currency, limit, aggregate)
		},
	)
	if err != nil {
		return nil, err
	}

	return prices.([]thirdparty.HistoricalPrice), nil
}

func (pm *Manager) FetchTokenMarketValues(symbols []string, currency string) (map[string]thirdparty.TokenMarketValues, error) {
	marketValues, err := pm.makeCall(
		func() (any, error) {
			return pm.main.FetchTokenMarketValues(symbols, currency)
		},
		func() (any, error) {
			return pm.fallback.FetchTokenMarketValues(symbols, currency)
		},
	)
	if err != nil {
		return nil, err
	}

	return marketValues.(map[string]thirdparty.TokenMarketValues), nil
}

func (pm *Manager) FetchTokenDetails(symbols []string) (map[string]thirdparty.TokenDetails, error) {
	tokenDetails, err := pm.makeCall(
		func() (any, error) {
			return pm.main.FetchTokenDetails(symbols)
		},
		func() (any, error) {
			return pm.fallback.FetchTokenDetails(symbols)
		},
	)
	if err != nil {
		return nil, err
	}

	return tokenDetails.(map[string]thirdparty.TokenDetails), nil
}

func (pm *Manager) FetchPrice(symbol string, currency string) (float64, error) {
	symbols := [1]string{symbol}
	currencies := [1]string{currency}

	prices, err := pm.FetchPrices(symbols[:], currencies[:])

	if err != nil {
		return 0, err
	}

	return prices[symbol][currency], nil
}

func (pm *Manager) FetchPrices(symbols []string, currencies []string) (map[string]map[string]float64, error) {
	result, err := pm.makeCall(
		func() (any, error) {
			return pm.main.FetchPrices(symbols, currencies)
		},
		func() (any, error) {
			return pm.fallback.FetchPrices(symbols, currencies)
		},
	)

	if err != nil {
		return nil, err
	}

	prices := result.(map[string]map[string]float64)
	pm.updatePriceCache(prices)
	return prices, nil
}

func (pm *Manager) getCachedPricesFor(symbols []string, currencies []string) DataPerTokenAndCurrency {
	prices := make(DataPerTokenAndCurrency)

	for _, symbol := range symbols {
		prices[symbol] = make(map[string]DataPoint)
		for _, currency := range currencies {
			prices[symbol][currency] = pm.priceCache[symbol][currency]
		}
	}

	return prices
}

func (pm *Manager) updatePriceCache(prices map[string]map[string]float64) {
	pm.priceCacheLock.Lock()
	defer pm.priceCacheLock.Unlock()

	for token, pricesPerCurrency := range prices {
		_, present := pm.priceCache[token]
		if !present {
			pm.priceCache[token] = make(map[string]DataPoint)
		}
		for currency, price := range pricesPerCurrency {
			pm.priceCache[token][currency] = DataPoint{
				Price:     price,
				UpdatedAt: time.Now().Unix(),
			}
		}
	}
}

func (pm *Manager) GetCachedPrices() DataPerTokenAndCurrency {
	pm.priceCacheLock.RLock()
	defer pm.priceCacheLock.RUnlock()

	return pm.priceCache
}

// Return cached price if present in cache and age is less than maxAgeInSeconds. Fetch otherwise.
func (pm *Manager) GetOrFetchPrices(symbols []string, currencies []string, maxAgeInSeconds int64) (DataPerTokenAndCurrency, error) {
	symbolsToFetchMap := make(map[string]bool)
	symbolsToFetch := make([]string, 0, len(symbols))

	now := time.Now().Unix()

	for _, symbol := range symbols {
		tokenPriceCache, ok := pm.GetCachedPrices()[symbol]
		if !ok {
			if !symbolsToFetchMap[symbol] {
				symbolsToFetchMap[symbol] = true
				symbolsToFetch = append(symbolsToFetch, symbol)
			}
			continue
		}
		for _, currency := range currencies {
			if now-tokenPriceCache[currency].UpdatedAt > maxAgeInSeconds {
				if !symbolsToFetchMap[symbol] {
					symbolsToFetchMap[symbol] = true
					symbolsToFetch = append(symbolsToFetch, symbol)
				}
				break
			}
		}
	}

	if len(symbolsToFetch) > 0 {
		_, err := pm.FetchPrices(symbolsToFetch, currencies)
		if err != nil {
			return nil, err
		}
	}

	prices := pm.getCachedPricesFor(symbols, currencies)

	return prices, nil
}
