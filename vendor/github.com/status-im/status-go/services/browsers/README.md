Browsers Service
================

Browser service provides read/write API for browser object.

To enable include browsers config part and add `browsers` to APIModules:


```json
{
  "BrowsersConfig": {
    "Enabled": true,
  },
  APIModules: "browsers"
}
```

API
---

Enabling service will expose three additional methods:

#### browsers_addBrowser

Stores browser in the database.
All fields are specified below:

```json
{
  "browser-id": "1",
  "name": "first",
  "timestamp": 10,
  "dapp?": true,
  "history-index": 1,
  "history": [
    "one",
    "two"
  ]
}
```

#### browsers_getBrowsers

Reads all browsers, returns in the format specified above. List is sorted by timestamp.

#### browsers_deleteBrowser

Delete browser from database. Accepts browser `id`.