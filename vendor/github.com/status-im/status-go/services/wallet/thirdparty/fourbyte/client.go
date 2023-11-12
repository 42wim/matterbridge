package fourbyte

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/status-im/status-go/services/wallet/thirdparty"
)

type Signature struct {
	ID   int    `json:"id"`
	Text string `json:"text_signature"`
}

type ByID []Signature

func (s ByID) Len() int           { return len(s) }
func (s ByID) Less(i, j int) bool { return s[i].ID > s[j].ID }
func (s ByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type SignatureList struct {
	Count   int         `json:"count"`
	Results []Signature `json:"results"`
}

type Client struct {
	Client *http.Client
	URL    string
}

func NewClient() *Client {
	return &Client{Client: &http.Client{Timeout: time.Minute}, URL: "https://www.4byte.directory"}
}

func (c *Client) DoQuery(url string) (*http.Response, error) {
	resp, err := c.Client.Get(url)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Run(data string) (*thirdparty.DataParsed, error) {
	if len(data) < 10 || !strings.HasPrefix(data, "0x") {
		return nil, errors.New("input is badly formatted")
	}
	methodSigData := data[2:10]
	url := fmt.Sprintf("%s/api/v1/signatures/?hex_signature=%s", c.URL, methodSigData)
	resp, err := c.DoQuery(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var signatures SignatureList
	err = json.Unmarshal(body, &signatures)
	if err != nil {
		return nil, err
	}
	if signatures.Count == 0 {
		return nil, err
	}
	rgx := regexp.MustCompile(`\((.*?)\)`)
	results := signatures.Results
	sort.Sort(ByID(results))
	for _, signature := range results {
		id := fmt.Sprintf("0x%x", signature.ID)
		name := strings.Split(signature.Text, "(")[0]
		rs := rgx.FindStringSubmatch(signature.Text)
		inputsMapString := make(map[string]string)
		if len(rs[1]) > 0 {
			inputs := make([]string, 0)
			rawInputs := strings.Split(rs[1], ",")
			for index, typ := range rawInputs {
				if index == len(rawInputs)-1 && typ == "bytes" {
					continue
				}
				inputs = append(inputs, fmt.Sprintf("{\"name\":\"%d\",\"type\":\"%s\"}", index, typ))
			}
			functionABI := fmt.Sprintf("[{\"constant\":true,\"inputs\":[%s],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\", \"name\": \"%s\"}], ", strings.Join(inputs, ","), name)
			contractABI, err := abi.JSON(strings.NewReader(functionABI))
			if err != nil {
				continue
			}
			method := contractABI.Methods[name]
			inputsMap := make(map[string]interface{})
			if err := method.Inputs.UnpackIntoMap(inputsMap, []byte(data[10:])); err != nil {
				continue
			}

			for key, value := range inputsMap {
				inputsMapString[key] = fmt.Sprintf("%v", value)
			}
		}

		return &thirdparty.DataParsed{
			Name:      name,
			ID:        id,
			Signature: signature.Text,
			Inputs:    inputsMapString,
		}, nil
	}

	return nil, errors.New("couldn't find a corresponding signature")
}
