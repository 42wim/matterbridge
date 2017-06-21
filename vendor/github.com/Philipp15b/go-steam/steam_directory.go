package steam

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Philipp15b/go-steam/netutil"
)

// Load initial server list from Steam Directory Web API.
// Call InitializeSteamDirectory() before Connect() to use
// steam directory server list instead of static one.
func InitializeSteamDirectory() error {
	return steamDirectoryCache.Initialize()
}

var steamDirectoryCache *steamDirectory = &steamDirectory{}

type steamDirectory struct {
	sync.RWMutex
	servers       []string
	isInitialized bool
}

// Get server list from steam directory and save it for later
func (sd *steamDirectory) Initialize() error {
	sd.Lock()
	defer sd.Unlock()
	client := new(http.Client)
	resp, err := client.Get(fmt.Sprintf("https://api.steampowered.com/ISteamDirectory/GetCMList/v1/?cellId=0"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	r := struct {
		Response struct {
			ServerList []string
			Result     uint32
			Message    string
		}
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Response.Result != 1 {
		return fmt.Errorf("Failed to get steam directory, result: %v, message: %v\n", r.Response.Result, r.Response.Message)
	}
	if len(r.Response.ServerList) == 0 {
		return fmt.Errorf("Steam returned zero servers for steam directory request\n")
	}
	sd.servers = r.Response.ServerList
	sd.isInitialized = true
	return nil
}

func (sd *steamDirectory) GetRandomCM() *netutil.PortAddr {
	sd.RLock()
	defer sd.RUnlock()
	if !sd.isInitialized {
		panic("steam directory is not initialized")
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	addr := netutil.ParsePortAddr(sd.servers[rng.Int31n(int32(len(sd.servers)))])
	return addr
}

func (sd *steamDirectory) IsInitialized() bool {
	sd.RLock()
	defer sd.RUnlock()
	isInitialized := sd.isInitialized
	return isInitialized
}
