package cryptocompare

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/status-im/status-go/services/wallet/thirdparty"
	"github.com/status-im/status-go/services/wallet/thirdparty/utils"
)

const baseURL = "https://min-api.cryptocompare.com"

type HistoricalPricesContainer struct {
	Aggregated     bool                         `json:"Aggregated"`
	TimeFrom       int64                        `json:"TimeFrom"`
	TimeTo         int64                        `json:"TimeTo"`
	HistoricalData []thirdparty.HistoricalPrice `json:"Data"`
}

type HistoricalPricesData struct {
	Data HistoricalPricesContainer `json:"Data"`
}

type TokenDetailsContainer struct {
	Data map[string]thirdparty.TokenDetails `json:"Data"`
}

type MarketValuesContainer struct {
	Raw map[string]map[string]thirdparty.TokenMarketValues `json:"Raw"`
}

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{client: &http.Client{Timeout: time.Minute}}
}

func (c *Client) DoQuery(url string) (*http.Response, error) {
	resp, err := c.client.Get(url)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) FetchPrices(symbols []string, currencies []string) (map[string]map[string]float64, error) {
	chunks := utils.ChunkSymbols(symbols, 60)
	result := make(map[string]map[string]float64)
	realCurrencies := utils.RenameSymbols(currencies)
	for _, smbls := range chunks {
		realSymbols := utils.RenameSymbols(smbls)
		url := fmt.Sprintf("%s/data/pricemulti?fsyms=%s&tsyms=%s&extraParams=Status.im", baseURL, strings.Join(realSymbols, ","), strings.Join(realCurrencies, ","))
		resp, err := c.DoQuery(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		prices := make(map[string]map[string]float64)
		err = json.Unmarshal(body, &prices)
		if err != nil {
			return nil, err
		}

		for _, symbol := range smbls {
			result[symbol] = map[string]float64{}
			for _, currency := range currencies {
				result[symbol][currency] = prices[utils.GetRealSymbol(symbol)][utils.GetRealSymbol(currency)]
			}
		}
	}
	return result, nil
}

func (c *Client) FetchTokenDetails(symbols []string) (map[string]thirdparty.TokenDetails, error) {
	url := fmt.Sprintf("%s/data/all/coinlist", baseURL)
	resp, err := c.DoQuery(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	container := TokenDetailsContainer{}
	err = json.Unmarshal(body, &container)
	if err != nil {
		return nil, err
	}

	tokenDetails := make(map[string]thirdparty.TokenDetails)

	for _, symbol := range symbols {
		tokenDetails[symbol] = container.Data[utils.GetRealSymbol(symbol)]
	}

	return tokenDetails, nil
}

func (c *Client) FetchTokenMarketValues(symbols []string, currency string) (map[string]thirdparty.TokenMarketValues, error) {
	chunks := utils.ChunkSymbols(symbols)
	realCurrency := utils.GetRealSymbol(currency)
	item := map[string]thirdparty.TokenMarketValues{}
	for _, smbls := range chunks {
		realSymbols := utils.RenameSymbols(smbls)
		url := fmt.Sprintf("%s/data/pricemultifull?fsyms=%s&tsyms=%s&extraParams=Status.im", baseURL, strings.Join(realSymbols, ","), realCurrency)
		resp, err := c.DoQuery(url)
		if err != nil {
			return item, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return item, err
		}

		container := MarketValuesContainer{}
		err = json.Unmarshal(body, &container)
		if err != nil {
			return item, err
		}

		for _, symbol := range smbls {
			item[symbol] = container.Raw[utils.GetRealSymbol(symbol)][utils.GetRealSymbol(currency)]
		}
	}
	return item, nil
}

func (c *Client) FetchHistoricalHourlyPrices(symbol string, currency string, limit int, aggregate int) ([]thirdparty.HistoricalPrice, error) {
	item := []thirdparty.HistoricalPrice{}

	url := fmt.Sprintf("%s/data/v2/histohour?fsym=%s&tsym=%s&aggregate=%d&limit=%d&extraParams=Status.im", baseURL, utils.GetRealSymbol(symbol), currency, aggregate, limit)
	resp, err := c.DoQuery(url)
	if err != nil {
		return item, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return item, err
	}

	container := HistoricalPricesData{}
	err = json.Unmarshal(body, &container)
	if err != nil {
		return item, err
	}

	item = container.Data.HistoricalData

	return item, nil
}

func (c *Client) FetchHistoricalDailyPrices(symbol string, currency string, limit int, allData bool, aggregate int) ([]thirdparty.HistoricalPrice, error) {
	item := []thirdparty.HistoricalPrice{}

	url := fmt.Sprintf("%s/data/v2/histoday?fsym=%s&tsym=%s&aggregate=%d&limit=%d&allData=%v&extraParams=Status.im", baseURL, utils.GetRealSymbol(symbol), currency, aggregate, limit, allData)
	resp, err := c.DoQuery(url)
	if err != nil {
		return item, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return item, err
	}

	container := HistoricalPricesData{}
	err = json.Unmarshal(body, &container)
	if err != nil {
		return item, err
	}

	item = container.Data.HistoricalData

	return item, nil
}
