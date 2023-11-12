package network

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
)

var SepoliaChainIDs = []uint64{11155111, 421614, 11155420}

var GoerliChainIDs = []uint64{5, 421613, 420}

type CombinedNetwork struct {
	Prod *params.Network
	Test *params.Network
}

const baseQuery = "SELECT chain_id, chain_name, rpc_url, original_rpc_url, fallback_url, original_fallback_url, block_explorer_url, icon_url, native_currency_name, native_currency_symbol, native_currency_decimals, is_test, layer, enabled, chain_color, short_name, related_chain_id FROM networks"

func newNetworksQuery() *networksQuery {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(baseQuery)
	return &networksQuery{buf: buf}
}

type networksQuery struct {
	buf   *bytes.Buffer
	args  []interface{}
	added bool
}

func (nq *networksQuery) andOrWhere() {
	if nq.added {
		nq.buf.WriteString(" AND")
	} else {
		nq.buf.WriteString(" WHERE")
	}
}

func (nq *networksQuery) filterEnabled(enabled bool) *networksQuery {
	nq.andOrWhere()
	nq.added = true
	nq.buf.WriteString(" enabled = ?")
	nq.args = append(nq.args, enabled)
	return nq
}

func (nq *networksQuery) filterChainID(chainID uint64) *networksQuery {
	nq.andOrWhere()
	nq.added = true
	nq.buf.WriteString(" chain_id = ?")
	nq.args = append(nq.args, chainID)
	return nq
}

func (nq *networksQuery) exec(db *sql.DB) ([]*params.Network, error) {
	rows, err := db.Query(nq.buf.String(), nq.args...)
	if err != nil {
		return nil, err
	}
	var res []*params.Network
	defer rows.Close()
	for rows.Next() {
		network := params.Network{}
		err := rows.Scan(
			&network.ChainID, &network.ChainName, &network.RPCURL, &network.OriginalRPCURL, &network.FallbackURL, &network.OriginalFallbackURL,
			&network.BlockExplorerURL, &network.IconURL, &network.NativeCurrencyName, &network.NativeCurrencySymbol,
			&network.NativeCurrencyDecimals, &network.IsTest, &network.Layer, &network.Enabled, &network.ChainColor, &network.ShortName,
			&network.RelatedChainID,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &network)
	}

	return res, err
}

type Manager struct {
	db                 *sql.DB
	configuredNetworks []params.Network
	accountsDB         *accounts.Database
}

func NewManager(db *sql.DB) *Manager {
	accountsDB, err := accounts.NewDB(db)
	if err != nil {
		return nil
	}
	return &Manager{
		db:         db,
		accountsDB: accountsDB,
	}
}

func find(chainID uint64, networks []params.Network) int {
	for i := range networks {
		if networks[i].ChainID == chainID {
			return i
		}
	}
	return -1
}

func (nm *Manager) Init(networks []params.Network) error {
	if networks == nil {
		return nil
	}
	nm.configuredNetworks = networks

	var errors string
	currentNetworks, _ := nm.Get(false)

	// Delete networks which are not supported any more
	for i := range currentNetworks {
		if find(currentNetworks[i].ChainID, networks) == -1 {
			err := nm.Delete(currentNetworks[i].ChainID)
			if err != nil {
				errors += fmt.Sprintf("error deleting network with ChainID: %d, %s", currentNetworks[i].ChainID, err.Error())
			}
		}
	}

	// Add new networks and update related chain id for the old ones
	for i := range networks {
		found := false
		networks[i].OriginalRPCURL = networks[i].RPCURL
		networks[i].OriginalFallbackURL = networks[i].FallbackURL

		for j := range currentNetworks {
			if currentNetworks[j].ChainID == networks[i].ChainID {
				found = true
				if currentNetworks[j].RelatedChainID != networks[i].RelatedChainID {
					// Update fallback_url if it's different
					err := nm.UpdateRelatedChainID(currentNetworks[j].ChainID, networks[i].RelatedChainID)
					if err != nil {
						errors += fmt.Sprintf("error updating network fallback_url for ChainID: %d, %s", currentNetworks[j].ChainID, err.Error())
					}
				}

				if networks[i].OriginalRPCURL != currentNetworks[j].OriginalRPCURL && currentNetworks[j].RPCURL == currentNetworks[j].OriginalRPCURL {
					err := nm.updateRPCURL(networks[i].ChainID, networks[i].OriginalRPCURL)
					if err != nil {
						errors += fmt.Sprintf("error updating rpc url for ChainID: %d, %s", currentNetworks[j].ChainID, err.Error())
					}
				}

				if networks[i].OriginalFallbackURL != currentNetworks[j].OriginalFallbackURL && currentNetworks[j].FallbackURL == currentNetworks[j].OriginalFallbackURL {
					err := nm.updateFallbackURL(networks[i].ChainID, networks[i].OriginalFallbackURL)
					if err != nil {
						errors += fmt.Sprintf("error updating rpc url for ChainID: %d, %s", currentNetworks[j].ChainID, err.Error())
					}
				}

				err := nm.updateOriginalURLs(networks[i].ChainID, networks[i].OriginalRPCURL, networks[i].OriginalFallbackURL)
				if err != nil {
					errors += fmt.Sprintf("error updating network original url for ChainID: %d, %s", currentNetworks[j].ChainID, err.Error())
				}
				break
			}
		}

		if !found {
			// Insert new network
			err := nm.Upsert(&networks[i])
			if err != nil {
				errors += fmt.Sprintf("error inserting network with ChainID: %d, %s", networks[i].ChainID, err.Error())
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(errors)
	}

	return nil
}

func (nm *Manager) Upsert(network *params.Network) error {
	_, err := nm.db.Exec(
		"INSERT OR REPLACE INTO networks (chain_id, chain_name, rpc_url, original_rpc_url, fallback_url, original_fallback_url, block_explorer_url, icon_url, native_currency_name, native_currency_symbol, native_currency_decimals, is_test, layer, enabled, chain_color, short_name, related_chain_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		network.ChainID, network.ChainName, network.RPCURL, network.OriginalRPCURL, network.FallbackURL, network.OriginalFallbackURL, network.BlockExplorerURL, network.IconURL,
		network.NativeCurrencyName, network.NativeCurrencySymbol, network.NativeCurrencyDecimals,
		network.IsTest, network.Layer, network.Enabled, network.ChainColor, network.ShortName,
		network.RelatedChainID,
	)
	return err
}

func (nm *Manager) Delete(chainID uint64) error {
	_, err := nm.db.Exec("DELETE FROM networks WHERE chain_id = ?", chainID)
	return err
}

func (nm *Manager) UpdateRelatedChainID(chainID uint64, relatedChainID uint64) error {
	_, err := nm.db.Exec(`UPDATE networks SET related_chain_id = ? WHERE chain_id = ?`, relatedChainID, chainID)
	return err
}

func (nm *Manager) updateRPCURL(chainID uint64, rpcURL string) error {
	_, err := nm.db.Exec(`UPDATE networks SET rpc_url = ? WHERE chain_id = ?`, rpcURL, chainID)
	return err
}

func (nm *Manager) updateFallbackURL(chainID uint64, fallbackURL string) error {
	_, err := nm.db.Exec(`UPDATE networks SET fallback_url = ? WHERE chain_id = ?`, fallbackURL, chainID)
	return err
}

func (nm *Manager) updateOriginalURLs(chainID uint64, originalRPCURL, OriginalFallbackURL string) error {
	_, err := nm.db.Exec(`UPDATE networks SET original_rpc_url = ?, original_fallback_url = ?  WHERE chain_id = ?`, originalRPCURL, OriginalFallbackURL, chainID)
	return err
}

func (nm *Manager) Find(chainID uint64) *params.Network {
	networks, err := newNetworksQuery().filterChainID(chainID).exec(nm.db)
	if len(networks) != 1 || err != nil {
		return nil
	}
	return networks[0]
}

func (nm *Manager) GetAll() ([]*params.Network, error) {
	query := newNetworksQuery()
	return query.exec(nm.db)
}

func (nm *Manager) Get(onlyEnabled bool) ([]*params.Network, error) {
	isSepoliaEnabled, err := nm.accountsDB.GetIsSepoliaEnabled()
	if err != nil {
		return nil, err
	}

	query := newNetworksQuery()
	if onlyEnabled {
		query.filterEnabled(true)
	}

	networks, err := query.exec(nm.db)
	if err != nil {
		return nil, err
	}

	var results []*params.Network
	for _, network := range networks {
		if !isSepoliaEnabled {
			found := false
			for _, chainID := range SepoliaChainIDs {
				if network.ChainID == chainID {
					found = true
					break
				}
			}
			if found {
				continue
			}
		}

		if isSepoliaEnabled {
			found := false

			for _, chainID := range GoerliChainIDs {
				if network.ChainID == chainID {
					found = true
					break
				}
			}
			if found {
				continue
			}
		}
		results = append(results, network)
	}

	return results, nil
}

func (nm *Manager) GetCombinedNetworks() ([]*CombinedNetwork, error) {
	networks, err := nm.Get(false)
	if err != nil {
		return nil, err
	}
	var combinedNetworks []*CombinedNetwork
	for _, network := range networks {
		found := false
		for _, n := range combinedNetworks {
			if (n.Test != nil && (network.ChainID == n.Test.RelatedChainID || n.Test.ChainID == network.RelatedChainID)) || (n.Prod != nil && (network.ChainID == n.Prod.RelatedChainID || n.Prod.ChainID == network.RelatedChainID)) {
				found = true
				if network.IsTest {
					n.Test = network
					break
				} else {
					n.Prod = network
					break
				}
			}
		}

		if found {
			continue
		}

		newCombined := &CombinedNetwork{}
		if network.IsTest {
			newCombined.Test = network
		} else {
			newCombined.Prod = network
		}
		combinedNetworks = append(combinedNetworks, newCombined)
	}

	return combinedNetworks, nil
}

func (nm *Manager) GetConfiguredNetworks() []params.Network {
	return nm.configuredNetworks
}

func (nm *Manager) GetTestNetworksEnabled() (result bool, err error) {
	return nm.accountsDB.GetTestNetworksEnabled()
}
