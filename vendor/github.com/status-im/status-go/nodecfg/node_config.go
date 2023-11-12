package nodecfg

import (
	"context"
	"database/sql"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/sqlite"
)

const StaticNodes = "static"
const BootNodes = "boot"
const TrustedMailServers = "trusted_mailserver"
const PushNotificationsServers = "pushnotification"
const RendezvousNodes = "rendezvous"
const DiscV5BootstrapNodes = "discV5boot"
const WakuNodes = "waku"

func nodeConfigWasMigrated(tx *sql.Tx) (migrated bool, err error) {
	row := tx.QueryRow("SELECT exists(SELECT 1 FROM node_config)")
	switch err := row.Scan(&migrated); err {
	case sql.ErrNoRows, nil:
		return migrated, nil
	default:
		return migrated, err
	}
}

type insertFn func(tx *sql.Tx, c *params.NodeConfig) error

func insertNodeConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
	INSERT OR REPLACE INTO node_config (
		network_id, data_dir, keystore_dir, node_key, no_discovery, rendezvous,
		listen_addr, advertise_addr, name, version, api_modules, tls_enabled,
		max_peers, max_pending_peers, enable_status_service, enable_ntp_sync,
		bridge_enabled, wallet_enabled, local_notifications_enabled,
		browser_enabled, permissions_enabled, mailservers_enabled,
		swarm_enabled, mailserver_registry_address, web3provider_enabled, synthetic_id
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, 'id'
	)`,
		c.NetworkID, c.DataDir, c.KeyStoreDir, c.NodeKey, c.NoDiscovery, c.Rendezvous,
		c.ListenAddr, c.AdvertiseAddr, c.Name, c.Version, c.APIModules,
		c.TLSEnabled, c.MaxPeers, c.MaxPendingPeers,
		c.EnableStatusService, true,
		c.BridgeConfig.Enabled, c.WalletConfig.Enabled, c.LocalNotificationsConfig.Enabled,
		c.BrowsersConfig.Enabled, c.PermissionsConfig.Enabled, c.MailserversConfig.Enabled,
		c.SwarmConfig.Enabled, c.MailServerRegistryAddress, c.Web3ProviderConfig.Enabled,
	)
	return err
}

func insertHTTPConfig(tx *sql.Tx, c *params.NodeConfig) error {
	if _, err := tx.Exec(`INSERT OR REPLACE INTO http_config (enabled, host, port, synthetic_id) VALUES (?, ?, ?, 'id')`, c.HTTPEnabled, c.HTTPHost, c.HTTPPort); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM http_virtual_hosts WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for _, httpVirtualHost := range c.HTTPVirtualHosts {
		if _, err := tx.Exec(`INSERT OR REPLACE INTO http_virtual_hosts (host, synthetic_id) VALUES (?, 'id')`, httpVirtualHost); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`DELETE FROM http_cors WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for _, httpCors := range c.HTTPCors {
		if _, err := tx.Exec(`INSERT OR REPLACE INTO http_cors (cors, synthetic_id) VALUES (?, 'id')`, httpCors); err != nil {
			return err
		}
	}

	return nil
}

func insertLogConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
	INSERT OR REPLACE INTO log_config (
		enabled, mobile_system, log_dir, log_level, max_backups, max_size,
		file, compress_rotated, log_to_stderr, synthetic_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'id'	)`,
		c.LogEnabled, c.LogMobileSystem, c.LogDir, c.LogLevel, c.LogMaxBackups, c.LogMaxSize,
		c.LogFile, c.LogCompressRotated, c.LogToStderr,
	)

	return err
}

func insertLightETHConfigTrustedNodes(tx *sql.Tx, c *params.NodeConfig) error {
	if _, err := tx.Exec(`DELETE FROM light_eth_trusted_nodes WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for _, node := range c.LightEthConfig.TrustedNodes {
		_, err := tx.Exec(`INSERT OR REPLACE INTO light_eth_trusted_nodes (node, synthetic_id) VALUES (?, 'id')`, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertRegisterTopics(tx *sql.Tx, c *params.NodeConfig) error {
	if _, err := tx.Exec(`DELETE FROM register_topics WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for _, topic := range c.RegisterTopics {
		_, err := tx.Exec(`INSERT OR REPLACE INTO register_topics (topic, synthetic_id) VALUES (?, 'id')`, topic)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertRequireTopics(tx *sql.Tx, c *params.NodeConfig) error {
	if _, err := tx.Exec(`DELETE FROM require_topics WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for topic, limits := range c.RequireTopics {
		_, err := tx.Exec(`INSERT OR REPLACE INTO require_topics (topic, min, max, synthetic_id) VALUES (?, ?, ?, 'id')`, topic, limits.Min, limits.Max)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertIPCConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`INSERT OR REPLACE INTO ipc_config (enabled, file, synthetic_id) VALUES (?, ?, 'id')`, c.IPCEnabled, c.IPCFile)
	return err
}

func insertClusterConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`INSERT OR REPLACE INTO cluster_config (enabled, fleet, synthetic_id) VALUES (?, ?, 'id')`, c.ClusterConfig.Enabled, c.ClusterConfig.Fleet)
	return err
}

func insertUpstreamConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`INSERT OR REPLACE INTO upstream_config (enabled, url, synthetic_id) VALUES (?, ?, 'id')`, c.UpstreamConfig.Enabled, c.UpstreamConfig.URL)
	return err
}

func insertLightETHConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`INSERT OR REPLACE INTO light_eth_config (enabled, database_cache, min_trusted_fraction, synthetic_id) VALUES (?, ?, ?, 'id')`, c.LightEthConfig.Enabled, c.LightEthConfig.DatabaseCache, c.LightEthConfig.MinTrustedFraction)
	return err
}

func insertShhExtConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
	INSERT OR REPLACE INTO shhext_config (
		pfs_enabled, backup_disabled_data_dir, installation_id, mailserver_confirmations, enable_connection_manager,
		enable_last_used_monitor, connection_target, request_delay, max_server_failures, max_message_delivery_attempts,
		whisper_cache_dir, disable_generic_discovery_topic, send_v1_messages, data_sync_enabled, verify_transaction_url,
		verify_ens_url, verify_ens_contract_address, verify_transaction_chain_id, anon_metrics_server_enabled,
		anon_metrics_send_id, anon_metrics_server_postgres_uri, bandwidth_stats_enabled, synthetic_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'id')`,
		c.ShhextConfig.PFSEnabled, c.ShhextConfig.BackupDisabledDataDir, c.ShhextConfig.InstallationID, c.ShhextConfig.MailServerConfirmations, c.ShhextConfig.EnableConnectionManager,
		c.ShhextConfig.EnableLastUsedMonitor, c.ShhextConfig.ConnectionTarget, c.ShhextConfig.RequestsDelay, c.ShhextConfig.MaxServerFailures, c.ShhextConfig.MaxMessageDeliveryAttempts,
		c.ShhextConfig.WhisperCacheDir, c.ShhextConfig.DisableGenericDiscoveryTopic, c.ShhextConfig.SendV1Messages, c.ShhextConfig.DataSyncEnabled, c.ShhextConfig.VerifyTransactionURL,
		c.ShhextConfig.VerifyENSURL, c.ShhextConfig.VerifyENSContractAddress, c.ShhextConfig.VerifyTransactionChainID, c.ShhextConfig.AnonMetricsServerEnabled,
		c.ShhextConfig.AnonMetricsSendID, c.ShhextConfig.AnonMetricsServerPostgresURI, c.ShhextConfig.BandwidthStatsEnabled)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM shhext_default_push_notification_servers WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for _, pushNotifServ := range c.ShhextConfig.DefaultPushNotificationsServers {
		hexpubk := hexutil.Encode(crypto.FromECDSAPub(pushNotifServ.PublicKey))
		_, err := tx.Exec(`INSERT OR REPLACE INTO shhext_default_push_notification_servers (public_key, synthetic_id) VALUES (?, 'id')`, hexpubk)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertTorrentConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
  INSERT OR REPLACE INTO torrent_config (
    enabled, port, data_dir, torrent_dir, synthetic_id
  ) VALUES (?, ?, ?, ?, 'id')`,
		c.TorrentConfig.Enabled, c.TorrentConfig.Port, c.TorrentConfig.DataDir, c.TorrentConfig.TorrentDir,
	)
	return err
}

func insertWakuV2Config(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
	INSERT OR REPLACE INTO wakuv2_config (
		enabled, host, port, keep_alive_interval, light_client, full_node, discovery_limit, data_dir,
		max_message_size, enable_confirmations, peer_exchange, enable_discv5, udp_port,  auto_update, synthetic_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'id')`,
		c.WakuV2Config.Enabled, c.WakuV2Config.Host, c.WakuV2Config.Port, c.WakuV2Config.KeepAliveInterval, c.WakuV2Config.LightClient, c.WakuV2Config.FullNode, c.WakuV2Config.DiscoveryLimit, c.WakuV2Config.DataDir,
		c.WakuV2Config.MaxMessageSize, c.WakuV2Config.EnableConfirmations, c.WakuV2Config.PeerExchange, c.WakuV2Config.EnableDiscV5, c.WakuV2Config.UDPPort, c.WakuV2Config.AutoUpdate,
	)
	if err != nil {
		return err
	}

	return setWakuV2CustomNodes(tx, c.WakuV2Config.CustomNodes)
}

func setWakuV2CustomNodes(tx *sql.Tx, customNodes map[string]string) error {
	if _, err := tx.Exec(`DELETE FROM wakuv2_custom_nodes WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for name, multiaddress := range customNodes {
		// NOTE: synthetic id is redundant, name is effectively the primary key
		_, err := tx.Exec(`INSERT OR REPLACE INTO wakuv2_custom_nodes (name, multiaddress, synthetic_id) VALUES (?, ?, 'id')`, name, multiaddress)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertWakuV2StoreConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
	UPDATE wakuv2_config
	SET enable_store = ?, store_capacity = ?, store_seconds = ?
	WHERE synthetic_id = 'id'`,
		c.WakuV2Config.EnableStore, c.WakuV2Config.StoreCapacity, c.WakuV2Config.StoreSeconds,
	)

	return err
}

func insertWakuV2ShardConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
	UPDATE wakuv2_config
	SET use_shard_default_topic = ?
	WHERE synthetic_id = 'id'`,
		c.WakuV2Config.UseShardAsDefaultTopic,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	UPDATE cluster_config
	SET cluster_id = ?
	WHERE synthetic_id = 'id'`,
		c.ClusterConfig.ClusterID,
	)

	return err
}

func insertWakuConfig(tx *sql.Tx, c *params.NodeConfig) error {
	_, err := tx.Exec(`
	INSERT OR REPLACE INTO waku_config (
		enabled, light_client, full_node, enable_mailserver, data_dir, minimum_pow, mailserver_password, mailserver_rate_limit, mailserver_data_retention,
		ttl, max_message_size, enable_rate_limiter, packet_rate_limit_ip, packet_rate_limit_peer_id, bytes_rate_limit_ip, bytes_rate_limit_peer_id,
		rate_limit_tolerance, bloom_filter_mode, enable_confirmations, synthetic_id
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'id')`,
		c.WakuConfig.Enabled, c.WakuConfig.LightClient, c.WakuConfig.FullNode, c.WakuConfig.EnableMailServer, c.WakuConfig.DataDir, c.WakuConfig.MinimumPoW,
		c.WakuConfig.MailServerPassword, c.WakuConfig.MailServerRateLimit, c.WakuConfig.MailServerDataRetention, c.WakuConfig.TTL, c.WakuConfig.MaxMessageSize,
		c.WakuConfig.EnableRateLimiter, c.WakuConfig.PacketRateLimitIP, c.WakuConfig.PacketRateLimitPeerID, c.WakuConfig.BytesRateLimitIP, c.WakuConfig.BytesRateLimitPeerID,
		c.WakuConfig.RateLimitTolerance, c.WakuConfig.BloomFilterMode, c.WakuConfig.EnableConfirmations,
	)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT OR REPLACE INTO waku_config_db_pg (enabled, uri, synthetic_id) VALUES (?, ?, 'id')`, c.WakuConfig.DatabaseConfig.PGConfig.Enabled, c.WakuConfig.DatabaseConfig.PGConfig.URI); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM waku_softblacklisted_peers WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	for _, peerID := range c.WakuConfig.SoftBlacklistedPeerIDs {
		_, err := tx.Exec(`INSERT OR REPLACE INTO waku_softblacklisted_peers (peer_id, synthetic_id) VALUES (?, 'id')`, peerID)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertPushNotificationsServerConfig(tx *sql.Tx, c *params.NodeConfig) error {
	hexPrivKey := ""
	if c.PushNotificationServerConfig.Identity != nil {
		hexPrivKey = hexutil.Encode(crypto.FromECDSA(c.PushNotificationServerConfig.Identity))
	}
	_, err := tx.Exec(`INSERT OR REPLACE INTO push_notifications_server_config (enabled, identity, gorush_url, synthetic_id) VALUES (?, ?, ?, 'id')`, c.PushNotificationServerConfig.Enabled, hexPrivKey, c.PushNotificationServerConfig.GorushURL)
	return err
}

func insertClusterConfigNodes(tx *sql.Tx, c *params.NodeConfig) error {
	if _, err := tx.Exec(`DELETE FROM cluster_nodes WHERE synthetic_id = 'id'`); err != nil {
		return err
	}

	nodeMap := make(map[string][]string)
	nodeMap[StaticNodes] = c.ClusterConfig.StaticNodes
	nodeMap[BootNodes] = c.ClusterConfig.BootNodes
	nodeMap[TrustedMailServers] = c.ClusterConfig.TrustedMailServers
	nodeMap[PushNotificationsServers] = c.ClusterConfig.PushNotificationsServers
	nodeMap[RendezvousNodes] = c.ClusterConfig.RendezvousNodes
	nodeMap[DiscV5BootstrapNodes] = c.ClusterConfig.DiscV5BootstrapNodes
	nodeMap[WakuNodes] = c.ClusterConfig.WakuNodes

	for nodeType, nodes := range nodeMap {
		for _, node := range nodes {
			_, err := tx.Exec(`INSERT OR REPLACE INTO cluster_nodes (node, type, synthetic_id) VALUES (?, ?, 'id')`, node, nodeType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// List of inserts to be executed when upgrading a node
// These INSERT queries should not be modified
func nodeConfigUpgradeInserts() []insertFn {
	return []insertFn{
		insertNodeConfig,
		insertHTTPConfig,
		insertIPCConfig,
		insertLogConfig,
		insertUpstreamConfig,
		insertClusterConfig,
		insertClusterConfigNodes,
		insertLightETHConfig,
		insertLightETHConfigTrustedNodes,
		insertRegisterTopics,
		insertRequireTopics,
		insertPushNotificationsServerConfig,
		insertShhExtConfig,
		insertWakuConfig,
		insertWakuV2Config,
	}
}

func nodeConfigNormalInserts() []insertFn {
	// WARNING: if you are modifying one of the node config tables
	// you need to edit `nodeConfigUpgradeInserts` to guarantee that
	// the selects being used there are not affected.

	return []insertFn{
		insertNodeConfig,
		insertHTTPConfig,
		insertIPCConfig,
		insertLogConfig,
		insertUpstreamConfig,
		insertClusterConfig,
		insertClusterConfigNodes,
		insertLightETHConfig,
		insertLightETHConfigTrustedNodes,
		insertRegisterTopics,
		insertRequireTopics,
		insertPushNotificationsServerConfig,
		insertShhExtConfig,
		insertWakuConfig,
		insertWakuV2Config,
		insertTorrentConfig,
		insertWakuV2StoreConfig,
		insertWakuV2ShardConfig,
	}
}

func execInsertFns(inFn []insertFn, tx *sql.Tx, c *params.NodeConfig) error {
	for _, fn := range inFn {
		err := fn(tx, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertNodeConfigUpgrade(tx *sql.Tx, c *params.NodeConfig) error {
	return execInsertFns(nodeConfigUpgradeInserts(), tx, c)
}

func SaveConfigWithTx(tx *sql.Tx, c *params.NodeConfig) error {
	insertFNs := nodeConfigNormalInserts()
	return execInsertFns(insertFNs, tx, c)
}

func SaveNodeConfig(db *sql.DB, c *params.NodeConfig) error {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	return SaveConfigWithTx(tx, c)
}

func migrateNodeConfig(tx *sql.Tx) error {
	nodecfg := &params.NodeConfig{}
	err := tx.QueryRow("SELECT node_config FROM settings WHERE synthetic_id = 'id'").Scan(&sqlite.JSONBlob{Data: nodecfg})

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err == sql.ErrNoRows {
		// Can't migrate because there's no data
		return nil
	}

	err = insertNodeConfigUpgrade(tx, nodecfg)
	if err != nil {
		return err
	}

	return nil
}

func loadNodeConfig(tx *sql.Tx) (*params.NodeConfig, error) {
	nodecfg := &params.NodeConfig{}

	err := tx.QueryRow(`
	SELECT
		network_id, data_dir, keystore_dir, node_key, no_discovery, rendezvous,
		listen_addr, advertise_addr, name, version, api_modules, tls_enabled, max_peers, max_pending_peers,
		enable_status_service, bridge_enabled, wallet_enabled, local_notifications_enabled,
		browser_enabled, permissions_enabled, mailservers_enabled, swarm_enabled,
		mailserver_registry_address, web3provider_enabled FROM node_config
		WHERE synthetic_id = 'id'
	`).Scan(
		&nodecfg.NetworkID, &nodecfg.DataDir, &nodecfg.KeyStoreDir, &nodecfg.NodeKey, &nodecfg.NoDiscovery, &nodecfg.Rendezvous,
		&nodecfg.ListenAddr, &nodecfg.AdvertiseAddr, &nodecfg.Name, &nodecfg.Version, &nodecfg.APIModules, &nodecfg.TLSEnabled, &nodecfg.MaxPeers, &nodecfg.MaxPendingPeers,
		&nodecfg.EnableStatusService, &nodecfg.BridgeConfig.Enabled, &nodecfg.WalletConfig.Enabled, &nodecfg.LocalNotificationsConfig.Enabled,
		&nodecfg.BrowsersConfig.Enabled, &nodecfg.PermissionsConfig.Enabled, &nodecfg.MailserversConfig.Enabled, &nodecfg.SwarmConfig.Enabled,
		&nodecfg.MailServerRegistryAddress, &nodecfg.Web3ProviderConfig.Enabled,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = tx.QueryRow(`SELECT enabled, host, port FROM http_config WHERE synthetic_id = 'id'`).Scan(&nodecfg.HTTPEnabled, &nodecfg.HTTPHost, &nodecfg.HTTPPort)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	rows, err := tx.Query("SELECT host FROM http_virtual_hosts WHERE synthetic_id = 'id' ORDER BY host ASC")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var host string
		err = rows.Scan(&host)
		if err != nil {
			return nil, err
		}
		nodecfg.HTTPVirtualHosts = append(nodecfg.HTTPVirtualHosts, host)
	}

	rows, err = tx.Query("SELECT cors FROM http_cors WHERE synthetic_id = 'id' ORDER BY cors ASC")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cors string
		err = rows.Scan(&cors)
		if err != nil {
			return nil, err
		}
		nodecfg.HTTPCors = append(nodecfg.HTTPCors, cors)
	}

	err = tx.QueryRow("SELECT enabled, file FROM ipc_config WHERE synthetic_id = 'id'").Scan(&nodecfg.IPCEnabled, &nodecfg.IPCFile)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = tx.QueryRow("SELECT enabled, mobile_system, log_dir, log_level, file, max_backups, max_size, compress_rotated, log_to_stderr FROM log_config WHERE synthetic_id = 'id'").Scan(
		&nodecfg.LogEnabled, &nodecfg.LogMobileSystem, &nodecfg.LogDir, &nodecfg.LogLevel, &nodecfg.LogFile, &nodecfg.LogMaxBackups, &nodecfg.LogMaxSize, &nodecfg.LogCompressRotated, &nodecfg.LogToStderr)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = tx.QueryRow("SELECT enabled, url FROM upstream_config WHERE synthetic_id = 'id'").Scan(&nodecfg.UpstreamConfig.Enabled, &nodecfg.UpstreamConfig.URL)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	rows, err = tx.Query(`SELECT
                chain_id, chain_name, rpc_url, block_explorer_url, icon_url, native_currency_name,
                native_currency_symbol, native_currency_decimals, is_test, layer, enabled, chain_color, short_name
        FROM networks ORDER BY chain_id ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n params.Network
		err = rows.Scan(&n.ChainID, &n.ChainName, &n.RPCURL, &n.BlockExplorerURL, &n.IconURL,
			&n.NativeCurrencyName, &n.NativeCurrencySymbol, &n.NativeCurrencyDecimals, &n.IsTest,
			&n.Layer, &n.Enabled, &n.ChainColor, &n.ShortName,
		)
		if err != nil {
			return nil, err
		}
		nodecfg.Networks = append(nodecfg.Networks, n)
	}

	err = tx.QueryRow("SELECT enabled, fleet, cluster_id FROM cluster_config WHERE synthetic_id = 'id'").Scan(&nodecfg.ClusterConfig.Enabled, &nodecfg.ClusterConfig.Fleet, &nodecfg.ClusterConfig.ClusterID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	nodeMap := make(map[string]*[]string)
	nodeMap[StaticNodes] = &nodecfg.ClusterConfig.StaticNodes
	nodeMap[BootNodes] = &nodecfg.ClusterConfig.BootNodes
	nodeMap[TrustedMailServers] = &nodecfg.ClusterConfig.TrustedMailServers
	nodeMap[PushNotificationsServers] = &nodecfg.ClusterConfig.PushNotificationsServers
	nodeMap[RendezvousNodes] = &nodecfg.ClusterConfig.RendezvousNodes
	nodeMap[WakuNodes] = &nodecfg.ClusterConfig.WakuNodes
	nodeMap[DiscV5BootstrapNodes] = &nodecfg.ClusterConfig.DiscV5BootstrapNodes
	rows, err = tx.Query(`SELECT node, type	FROM cluster_nodes WHERE synthetic_id = 'id' ORDER BY node ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var node string
		var nodeType string
		err = rows.Scan(&node, &nodeType)
		if err != nil {
			return nil, err
		}
		if nodeList, ok := nodeMap[nodeType]; ok {
			*nodeList = append(*nodeList, node)
		}
	}

	err = tx.QueryRow("SELECT enabled, database_cache, min_trusted_fraction FROM light_eth_config WHERE synthetic_id = 'id'").Scan(&nodecfg.LightEthConfig.Enabled, &nodecfg.LightEthConfig.DatabaseCache, &nodecfg.LightEthConfig.MinTrustedFraction)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	rows, err = tx.Query(`SELECT node FROM light_eth_trusted_nodes WHERE synthetic_id = 'id' ORDER BY node ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var node string
		err = rows.Scan(&node)
		if err != nil {
			return nil, err
		}
		nodecfg.LightEthConfig.TrustedNodes = append(nodecfg.LightEthConfig.TrustedNodes, node)
	}

	rows, err = tx.Query(`SELECT topic FROM register_topics WHERE synthetic_id = 'id' ORDER BY topic ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var topic discv5.Topic
		err = rows.Scan(&topic)
		if err != nil {
			return nil, err
		}
		nodecfg.RegisterTopics = append(nodecfg.RegisterTopics, topic)
	}

	rows, err = tx.Query(`SELECT topic, min, max FROM require_topics WHERE synthetic_id = 'id' ORDER BY topic ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	nodecfg.RequireTopics = make(map[discv5.Topic]params.Limits)
	for rows.Next() {
		var topic discv5.Topic
		var limit params.Limits
		err = rows.Scan(&topic, &limit.Min, &limit.Max)
		if err != nil {
			return nil, err
		}
		nodecfg.RequireTopics[topic] = limit
	}

	var pushNotifHexIdentity string
	err = tx.QueryRow("SELECT enabled, identity, gorush_url FROM push_notifications_server_config WHERE synthetic_id = 'id'").Scan(&nodecfg.PushNotificationServerConfig.Enabled, &pushNotifHexIdentity, &nodecfg.PushNotificationServerConfig.GorushURL)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if pushNotifHexIdentity != "" {
		b, err := hexutil.Decode(pushNotifHexIdentity)
		if err != nil {
			return nil, err
		}
		nodecfg.PushNotificationServerConfig.Identity, err = crypto.ToECDSA(b)
		if err != nil {
			return nil, err
		}
	}

	err = tx.QueryRow(`
	SELECT pfs_enabled, backup_disabled_data_dir, installation_id, mailserver_confirmations, enable_connection_manager,
	enable_last_used_monitor, connection_target, request_delay, max_server_failures, max_message_delivery_attempts,
	whisper_cache_dir, disable_generic_discovery_topic, send_v1_messages, data_sync_enabled, verify_transaction_url,
	verify_ens_url, verify_ens_contract_address, verify_transaction_chain_id, anon_metrics_server_enabled,
	anon_metrics_send_id, anon_metrics_server_postgres_uri, bandwidth_stats_enabled FROM shhext_config WHERE synthetic_id = 'id'
	`).Scan(
		&nodecfg.ShhextConfig.PFSEnabled, &nodecfg.ShhextConfig.BackupDisabledDataDir, &nodecfg.ShhextConfig.InstallationID, &nodecfg.ShhextConfig.MailServerConfirmations, &nodecfg.ShhextConfig.EnableConnectionManager,
		&nodecfg.ShhextConfig.EnableLastUsedMonitor, &nodecfg.ShhextConfig.ConnectionTarget, &nodecfg.ShhextConfig.RequestsDelay, &nodecfg.ShhextConfig.MaxServerFailures, &nodecfg.ShhextConfig.MaxMessageDeliveryAttempts,
		&nodecfg.ShhextConfig.WhisperCacheDir, &nodecfg.ShhextConfig.DisableGenericDiscoveryTopic, &nodecfg.ShhextConfig.SendV1Messages, &nodecfg.ShhextConfig.DataSyncEnabled, &nodecfg.ShhextConfig.VerifyTransactionURL,
		&nodecfg.ShhextConfig.VerifyENSURL, &nodecfg.ShhextConfig.VerifyENSContractAddress, &nodecfg.ShhextConfig.VerifyTransactionChainID, &nodecfg.ShhextConfig.AnonMetricsServerEnabled,
		&nodecfg.ShhextConfig.AnonMetricsSendID, &nodecfg.ShhextConfig.AnonMetricsServerPostgresURI, &nodecfg.ShhextConfig.BandwidthStatsEnabled,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	rows, err = tx.Query(`SELECT public_key FROM shhext_default_push_notification_servers WHERE synthetic_id = 'id' ORDER BY public_key ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pubKeyStr string
		err = rows.Scan(&pubKeyStr)
		if err != nil {
			return nil, err
		}

		if pubKeyStr != "" {
			b, err := hexutil.Decode(pubKeyStr)
			if err != nil {
				return nil, err
			}

			pubKey, err := crypto.UnmarshalPubkey(b)
			if err != nil {
				return nil, err
			}
			nodecfg.ShhextConfig.DefaultPushNotificationsServers = append(nodecfg.ShhextConfig.DefaultPushNotificationsServers, &params.PushNotificationServer{PublicKey: pubKey})
		}
	}

	err = tx.QueryRow(`
  SELECT enabled, port, data_dir, torrent_dir
  FROM torrent_config WHERE synthetic_id = 'id'
  `).Scan(
		&nodecfg.TorrentConfig.Enabled, &nodecfg.TorrentConfig.Port, &nodecfg.TorrentConfig.DataDir, &nodecfg.TorrentConfig.TorrentDir,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = tx.QueryRow(`
	SELECT enabled, host, port, keep_alive_interval, light_client, full_node, discovery_limit, data_dir,
	max_message_size, enable_confirmations, peer_exchange, enable_discv5, udp_port, auto_update,
	enable_store, store_capacity, store_seconds, use_shard_default_topic
	FROM wakuv2_config WHERE synthetic_id = 'id'
	`).Scan(
		&nodecfg.WakuV2Config.Enabled, &nodecfg.WakuV2Config.Host, &nodecfg.WakuV2Config.Port, &nodecfg.WakuV2Config.KeepAliveInterval, &nodecfg.WakuV2Config.LightClient, &nodecfg.WakuV2Config.FullNode,
		&nodecfg.WakuV2Config.DiscoveryLimit, &nodecfg.WakuV2Config.DataDir, &nodecfg.WakuV2Config.MaxMessageSize, &nodecfg.WakuV2Config.EnableConfirmations,
		&nodecfg.WakuV2Config.PeerExchange, &nodecfg.WakuV2Config.EnableDiscV5, &nodecfg.WakuV2Config.UDPPort, &nodecfg.WakuV2Config.AutoUpdate,
		&nodecfg.WakuV2Config.EnableStore, &nodecfg.WakuV2Config.StoreCapacity, &nodecfg.WakuV2Config.StoreSeconds, &nodecfg.WakuV2Config.UseShardAsDefaultTopic,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	rows, err = tx.Query(`SELECT name, multiaddress FROM wakuv2_custom_nodes WHERE synthetic_id = 'id' ORDER BY name ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	nodecfg.WakuV2Config.CustomNodes = make(map[string]string)
	for rows.Next() {
		var name string
		var multiaddress string
		err = rows.Scan(&name, &multiaddress)
		if err != nil {
			return nil, err
		}
		nodecfg.WakuV2Config.CustomNodes[name] = multiaddress
	}

	err = tx.QueryRow(`
	SELECT enabled, light_client, full_node, enable_mailserver, data_dir, minimum_pow, mailserver_password, mailserver_rate_limit, mailserver_data_retention,
	ttl, max_message_size, enable_rate_limiter, packet_rate_limit_ip, packet_rate_limit_peer_id, bytes_rate_limit_ip, bytes_rate_limit_peer_id,
	rate_limit_tolerance, bloom_filter_mode, enable_confirmations
	FROM waku_config WHERE synthetic_id = 'id'
	`).Scan(
		&nodecfg.WakuConfig.Enabled, &nodecfg.WakuConfig.LightClient, &nodecfg.WakuConfig.FullNode, &nodecfg.WakuConfig.EnableMailServer, &nodecfg.WakuConfig.DataDir, &nodecfg.WakuConfig.MinimumPoW,
		&nodecfg.WakuConfig.MailServerPassword, &nodecfg.WakuConfig.MailServerRateLimit, &nodecfg.WakuConfig.MailServerDataRetention, &nodecfg.WakuConfig.TTL, &nodecfg.WakuConfig.MaxMessageSize,
		&nodecfg.WakuConfig.EnableRateLimiter, &nodecfg.WakuConfig.PacketRateLimitIP, &nodecfg.WakuConfig.PacketRateLimitPeerID, &nodecfg.WakuConfig.BytesRateLimitIP, &nodecfg.WakuConfig.BytesRateLimitPeerID,
		&nodecfg.WakuConfig.RateLimitTolerance, &nodecfg.WakuConfig.BloomFilterMode, &nodecfg.WakuConfig.EnableConfirmations,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	err = tx.QueryRow("SELECT enabled, uri FROM waku_config_db_pg WHERE synthetic_id = 'id'").Scan(&nodecfg.WakuConfig.DatabaseConfig.PGConfig.Enabled, &nodecfg.WakuConfig.DatabaseConfig.PGConfig.URI)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	rows, err = tx.Query(`SELECT peer_id FROM waku_softblacklisted_peers WHERE synthetic_id = 'id' ORDER BY peer_id ASC`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var peerID string
		err = rows.Scan(&peerID)
		if err != nil {
			return nil, err
		}
		nodecfg.WakuConfig.SoftBlacklistedPeerIDs = append(nodecfg.WakuConfig.SoftBlacklistedPeerIDs, peerID)
	}

	return nodecfg, nil
}

func MigrateNodeConfig(db *sql.DB) error {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	migrated, err := nodeConfigWasMigrated(tx)
	if err != nil {
		return err
	}

	if !migrated {
		return migrateNodeConfig(tx)
	}

	return nil
}

func GetNodeConfigFromDB(db *sql.DB) (*params.NodeConfig, error) {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()

	return loadNodeConfig(tx)
}

func SetLightClient(db *sql.DB, enabled bool) error {
	_, err := db.Exec(`UPDATE wakuv2_config SET light_client = ?`, enabled)
	return err
}

func SetLogLevel(db *sql.DB, logLevel string) error {
	_, err := db.Exec(`UPDATE log_config SET log_level = ?`, logLevel)
	return err
}

func SetWakuV2CustomNodes(db *sql.DB, customNodes map[string]string) error {
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		// don't shadow original error
		_ = tx.Rollback()
	}()
	return setWakuV2CustomNodes(tx, customNodes)
}
