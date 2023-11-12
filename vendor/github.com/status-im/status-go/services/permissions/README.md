Dapps permissions service
=========================

To enable:

```json
{
  "PermissionsConfig": {
    "Enabled": true,
  },
  APIModules: "permissions"
}
```

API
---

#### permissions_addDappPermissions

Stores provided permissions for dapp. On update replaces previous version of the object.

```json
{
  "dapp": "first",
  "permissions": [
    "r",
    "x"
  ]
}
```

#### permissions_getDappPermissions

Returns all permissions for dapps. Order is not deterministic.

#### permissions_deleteDappPermissions

Delete dapp by a name.