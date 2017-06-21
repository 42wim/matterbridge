/*
Includes inventory types as used in the trade package
*/
package inventory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Philipp15b/go-steam/jsont"
	"strconv"
)

type GenericInventory map[uint32]map[uint64]*Inventory

func NewGenericInventory() GenericInventory {
	iMap := make(map[uint32]map[uint64]*Inventory)
	return GenericInventory(iMap)
}

// Get inventory for specified AppId and ContextId
func (i *GenericInventory) Get(appId uint32, contextId uint64) (*Inventory, error) {
	iMap := (map[uint32]map[uint64]*Inventory)(*i)
	iMap2, ok := iMap[appId]
	if !ok {
		return nil, fmt.Errorf("inventory for specified appId not found")
	}
	inv, ok := iMap2[contextId]
	if !ok {
		return nil, fmt.Errorf("inventory for specified contextId not found")
	}
	return inv, nil
}

func (i *GenericInventory) Add(appId uint32, contextId uint64, inv *Inventory) {
	iMap := (map[uint32]map[uint64]*Inventory)(*i)
	iMap2, ok := iMap[appId]
	if !ok {
		iMap2 = make(map[uint64]*Inventory)
		iMap[appId] = iMap2
	}
	iMap2[contextId] = inv
}

type Inventory struct {
	Items        Items        `json:"rgInventory"`
	Currencies   Currencies   `json:"rgCurrency"`
	Descriptions Descriptions `json:"rgDescriptions"`
	AppInfo      *AppInfo     `json:"rgAppInfo"`
}

// Items key is an AssetId
type Items map[string]*Item

func (i *Items) ToMap() map[string]*Item {
	return (map[string]*Item)(*i)
}

func (i *Items) Get(assetId uint64) (*Item, error) {
	iMap := (map[string]*Item)(*i)
	if item, ok := iMap[strconv.FormatUint(assetId, 10)]; ok {
		return item, nil
	}
	return nil, fmt.Errorf("item not found")
}

func (i *Items) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("[]")) {
		return nil
	}
	return json.Unmarshal(data, (*map[string]*Item)(i))
}

type Currencies map[string]*Currency

func (c *Currencies) ToMap() map[string]*Currency {
	return (map[string]*Currency)(*c)
}

func (c *Currencies) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("[]")) {
		return nil
	}
	return json.Unmarshal(data, (*map[string]*Currency)(c))
}

// Descriptions key format is %d_%d, first %d is ClassId, second is InstanceId
type Descriptions map[string]*Description

func (d *Descriptions) ToMap() map[string]*Description {
	return (map[string]*Description)(*d)
}

func (d *Descriptions) Get(classId uint64, instanceId uint64) (*Description, error) {
	dMap := (map[string]*Description)(*d)
	descId := fmt.Sprintf("%v_%v", classId, instanceId)
	if desc, ok := dMap[descId]; ok {
		return desc, nil
	}
	return nil, fmt.Errorf("description not found")
}

func (d *Descriptions) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("[]")) {
		return nil
	}
	return json.Unmarshal(data, (*map[string]*Description)(d))
}

type Item struct {
	Id         uint64 `json:",string"`
	ClassId    uint64 `json:",string"`
	InstanceId uint64 `json:",string"`
	Amount     uint64 `json:",string"`
	Pos        uint32
}

type Currency struct {
	Id         uint64 `json:",string"`
	ClassId    uint64 `json:",string"`
	IsCurrency bool   `json:"is_currency"`
	Pos        uint32
}

type Description struct {
	AppId      uint32 `json:",string"`
	ClassId    uint64 `json:",string"`
	InstanceId uint64 `json:",string"`

	IconUrl      string `json:"icon_url"`
	IconUrlLarge string `json:"icon_url_large"`
	IconDragUrl  string `json:"icon_drag_url"`

	Name           string
	MarketName     string `json:"market_name"`
	MarketHashName string `json:"market_hash_name"`

	// Colors in hex, for example `B2B2B2`
	NameColor       string `json:"name_color"`
	BackgroundColor string `json:"background_color"`

	Type string

	Tradable                  jsont.UintBool
	Marketable                jsont.UintBool
	Commodity                 jsont.UintBool
	MarketTradableRestriction uint32 `json:"market_tradable_restriction,string"`

	Descriptions DescriptionLines
	Actions      []*Action
	// Application-specific data, like "def_index" and "quality" for TF2
	AppData map[string]string
	Tags    []*Tag
}

type DescriptionLines []*DescriptionLine

func (d *DescriptionLines) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte(`""`)) {
		return nil
	}
	return json.Unmarshal(data, (*[]*DescriptionLine)(d))
}

type DescriptionLine struct {
	Value string
	Type  *string // Is `html` for HTML descriptions
	Color *string
}

type Action struct {
	Name string
	Link string
}

type AppInfo struct {
	AppId uint32
	Name  string
	Icon  string
	Link  string
}

type Tag struct {
	InternalName string `json:internal_name`
	Name         string
	Category     string
	CategoryName string `json:category_name`
}
