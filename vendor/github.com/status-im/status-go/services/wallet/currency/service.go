package currency

import (
	"context"
	"database/sql"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"github.com/status-im/status-go/services/wallet/market"
	"github.com/status-im/status-go/services/wallet/token"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

const (
	EventCurrencyTickUpdateFormat walletevent.EventType = "wallet-currency-tick-update-format"

	currencyFormatUpdateInterval = 1 * time.Hour
)

type Service struct {
	currency *Currency
	db       *DB

	tokenManager *token.Manager
	walletFeed   *event.Feed
	cancelFn     context.CancelFunc
}

func NewService(db *sql.DB, walletFeed *event.Feed, tokenManager *token.Manager, marketManager *market.Manager) *Service {
	return &Service{
		currency:     NewCurrency(marketManager),
		db:           NewCurrencyDB(db),
		tokenManager: tokenManager,
		walletFeed:   walletFeed,
	}
}

func (s *Service) Start() {
	// Update all fiat currency formats in cache
	fiatFormats, err := s.getAllFiatCurrencyFormats()

	if err == nil {
		_ = s.db.UpdateCachedFormats(fiatFormats)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel

	go func() {
		ticker := time.NewTicker(currencyFormatUpdateInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.walletFeed.Send(walletevent.Event{
					Type: EventCurrencyTickUpdateFormat,
				})
			}
		}
	}()
}

func (s *Service) Stop() {
	if s.cancelFn != nil {
		s.cancelFn()
	}
}

func (s *Service) GetCachedCurrencyFormats() (FormatPerSymbol, error) {
	return s.db.GetCachedFormats()
}

func (s *Service) FetchAllCurrencyFormats() (FormatPerSymbol, error) {
	// Only token prices can change, so we fetch those
	tokenFormats, err := s.fetchAllTokenCurrencyFormats()

	if err != nil {
		return nil, err
	}

	err = s.db.UpdateCachedFormats(tokenFormats)

	if err != nil {
		return nil, err
	}

	return s.GetCachedCurrencyFormats()
}

func (s *Service) getAllFiatCurrencyFormats() (FormatPerSymbol, error) {
	return GetFiatCurrencyFormats(GetAllFiatCurrencySymbols())
}

func (s *Service) fetchAllTokenCurrencyFormats() (FormatPerSymbol, error) {
	tokens, err := s.tokenManager.GetAllTokens()
	if err != nil {
		return nil, err
	}

	tokenPerSymbolMap := make(map[string]bool)
	tokenSymbols := make([]string, 0)
	for _, t := range tokens {
		symbol := t.Symbol
		if !tokenPerSymbolMap[symbol] {
			tokenPerSymbolMap[symbol] = true
			tokenSymbols = append(tokenSymbols, symbol)
		}
	}

	tokenFormats, err := s.currency.FetchTokenCurrencyFormats(tokenSymbols)
	if err != nil {
		return nil, err
	}
	gweiSymbol := "Gwei"
	tokenFormats[gweiSymbol] = Format{
		Symbol:              gweiSymbol,
		DisplayDecimals:     9,
		StripTrailingZeroes: true,
	}
	return tokenFormats, err
}
