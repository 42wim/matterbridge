package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-version"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/ens"
	"github.com/status-im/status-go/signal"
)

func NewAPI(ensService *ens.Service) *API {
	return &API{
		ensService: ensService,
		httpClient: &http.Client{Timeout: time.Minute},
	}
}

type API struct {
	ensService *ens.Service
	httpClient *http.Client
}

func (api *API) Check(ctx context.Context, chainID uint64, ens string, currentVersion string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		current, err := version.NewVersion(currentVersion)
		if err != nil {
			log.Error("invalid current version", "err", err)
			return
		}

		uri, err := api.ensService.API().ResourceURL(ctx, chainID, ens)
		if err != nil || uri.Host == "" {
			log.Error("can't get obtain the updates content hash url", "ens", ens)
			signal.SendUpdateAvailable(false, "", "")
			return
		}

		url := uri.Scheme + "://" + uri.Host + uri.Path
		versionURL := url + "VERSION"
		response, err := api.httpClient.Get(versionURL)
		if err != nil {
			log.Error("can't get content", zap.String("any", versionURL))
			signal.SendUpdateAvailable(false, "", "")
			return
		}

		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			log.Error(fmt.Sprintf("version verification response status error: %v", response.StatusCode))
			signal.SendUpdateAvailable(false, "", "")
			return
		}

		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Error("version verification body err", "err", err)
			signal.SendUpdateAvailable(false, "", "")
			return
		}

		c := make(map[string]interface{})
		err = json.Unmarshal(data, &c)
		if err != nil {
			log.Error("invalid json", "err", err)
			signal.SendUpdateAvailable(false, "", "")
			return
		}

		latestStr := ""
		switch c["version"].(type) {
		case string:
			latestStr = c["version"].(string)
		default:
			log.Error("invalid latest version", "val", c["version"])
			signal.SendUpdateAvailable(false, "", "")
			return
		}

		latest, err := version.NewVersion(latestStr)
		if err != nil {
			log.Error("invalid latest version", "err", err)
			signal.SendUpdateAvailable(false, "", "")
			return
		}

		signal.SendUpdateAvailable(latest.GreaterThan(current), latestStr, url)
	}()
}
