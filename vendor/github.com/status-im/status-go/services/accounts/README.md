Settings service
================

Settings service provides private API for storing all configuration for a selected account.

To enable:
1. Client must ensure that settings db is initialized in the api.Backend.
2. Add `settings` to APIModules in config.

API
---

### settings_saveConfig

#### Parameters

- `type`: `string` - configuratin type. if not unique error is raised.
- `conf`: `bytes` - raw json.

### settings_getConfig

#### Parameters

- `type`: string

#### Returns

- `conf` raw json

### settings_saveNodeConfig

Special case of the settings_saveConfig. In status-go we are using constant `node-config` as a type for node configuration.
Application depends on this value and will try to load it when node is started. This method is provided
in order to remove syncing mentioned constant between status-go and users.

#### Parameters

- `conf`: params.NodeConfig