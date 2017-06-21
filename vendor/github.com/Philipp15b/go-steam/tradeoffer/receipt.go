package tradeoffer

import (
	"encoding/json"
	"fmt"
	"github.com/Philipp15b/go-steam/economy/inventory"
	"regexp"
)

type TradeReceiptItem struct {
	AssetId   uint64 `json:"id,string"`
	AppId     uint32
	ContextId uint64
	Owner     uint64 `json:",string"`
	Pos       uint32
	inventory.Description
}

func parseTradeReceipt(data []byte) ([]*TradeReceiptItem, error) {
	reg := regexp.MustCompile("oItem =\\s+(.+?});")
	itemMatches := reg.FindAllSubmatch(data, -1)
	if itemMatches == nil {
		return nil, fmt.Errorf("items not found\n")
	}
	items := make([]*TradeReceiptItem, 0, len(itemMatches))
	for _, m := range itemMatches {
		item := new(TradeReceiptItem)
		err := json.Unmarshal(m[1], &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
