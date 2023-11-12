package thirdparty

type HistoricalPrice struct {
	Timestamp int64   `json:"time"`
	Value     float64 `json:"close"`
}

type TokenMarketValues struct {
	MKTCAP          float64 `json:"MKTCAP"`
	HIGHDAY         float64 `json:"HIGHDAY"`
	LOWDAY          float64 `json:"LOWDAY"`
	CHANGEPCTHOUR   float64 `json:"CHANGEPCTHOUR"`
	CHANGEPCTDAY    float64 `json:"CHANGEPCTDAY"`
	CHANGEPCT24HOUR float64 `json:"CHANGEPCT24HOUR"`
	CHANGE24HOUR    float64 `json:"CHANGE24HOUR"`
}

type TokenDetails struct {
	ID                   string  `json:"Id"`
	Name                 string  `json:"Name"`
	Symbol               string  `json:"Symbol"`
	Description          string  `json:"Description"`
	TotalCoinsMined      float64 `json:"TotalCoinsMined"`
	AssetLaunchDate      string  `json:"AssetLaunchDate"`
	AssetWhitepaperURL   string  `json:"AssetWhitepaperUrl"`
	AssetWebsiteURL      string  `json:"AssetWebsiteUrl"`
	BuiltOn              string  `json:"BuiltOn"`
	SmartContractAddress string  `json:"SmartContractAddress"`
}

type MarketDataProvider interface {
	FetchPrices(symbols []string, currencies []string) (map[string]map[string]float64, error)
	FetchHistoricalDailyPrices(symbol string, currency string, limit int, allData bool, aggregate int) ([]HistoricalPrice, error)
	FetchHistoricalHourlyPrices(symbol string, currency string, limit int, aggregate int) ([]HistoricalPrice, error)
	FetchTokenMarketValues(symbols []string, currency string) (map[string]TokenMarketValues, error)
	FetchTokenDetails(symbols []string) (map[string]TokenDetails, error)
}

type DataParsed struct {
	Name      string            `json:"name"`
	ID        string            `json:"id"`
	Inputs    map[string]string `json:"inputs"`
	Signature string            `json:"signature"`
}

type DecoderProvider interface {
	Run(data string) (*DataParsed, error)
}
