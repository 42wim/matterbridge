package inventory

import (
	"encoding/json"
	"fmt"
	"github.com/Philipp15b/go-steam/steamid"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

type InventoryApps map[string]*InventoryApp

func (i *InventoryApps) Get(appId uint32) (*InventoryApp, error) {
	iMap := (map[string]*InventoryApp)(*i)
	if inventoryApp, ok := iMap[strconv.FormatUint(uint64(appId), 10)]; ok {
		return inventoryApp, nil
	}
	return nil, fmt.Errorf("inventory app not found")
}

func (i *InventoryApps) ToMap() map[string]*InventoryApp {
	return (map[string]*InventoryApp)(*i)
}

type InventoryApp struct {
	AppId            uint32
	Name             string
	Icon             string
	Link             string
	AssetCount       uint32   `json:"asset_count"`
	InventoryLogo    string   `json:"inventory_logo"`
	TradePermissions string   `json:"trade_permissions"`
	Contexts         Contexts `json:"rgContexts"`
}

type Contexts map[string]*Context

func (c *Contexts) Get(contextId uint64) (*Context, error) {
	cMap := (map[string]*Context)(*c)
	if context, ok := cMap[strconv.FormatUint(contextId, 10)]; ok {
		return context, nil
	}
	return nil, fmt.Errorf("context not found")
}

func (c *Contexts) ToMap() map[string]*Context {
	return (map[string]*Context)(*c)
}

type Context struct {
	ContextId  uint64 `json:"id,string"`
	AssetCount uint32 `json:"asset_count"`
	Name       string
}

func GetInventoryApps(client *http.Client, steamId steamid.SteamId) (InventoryApps, error) {
	resp, err := http.Get("http://steamcommunity.com/profiles/" + steamId.ToString() + "/inventory/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	reg := regexp.MustCompile("var g_rgAppContextData = (.*?);")
	inventoryAppsMatches := reg.FindSubmatch(respBody)
	if inventoryAppsMatches == nil {
		return nil, fmt.Errorf("profile inventory not found in steam response")
	}
	var inventoryApps InventoryApps
	if err = json.Unmarshal(inventoryAppsMatches[1], &inventoryApps); err != nil {
		return nil, err
	}

	return inventoryApps, nil
}
