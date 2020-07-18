# msauth

## Introduction

Very simple package to authorize applications against [Microsoft identity platform].

It utilizes [v2.0 endpoint] so that it can authorize users using both personal (Microsoft) and organizational (Azure AD) account.

## Usage

### Device authorization grant

- [OAuth 2.0 device authorization grant flow]

```go
const (
	tenantID     = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	clientID     = "YYYYYYYY-YYYY-YYYY-YYYY-YYYYYYYYYYYY"
	tokenCachePath  = "token_cache.json"
)

var scopes = []string{"openid", "profile", "offline_access", "User.Read", "Files.Read"}

	ctx := context.Background()
	m := msauth.NewManager()
	m.LoadFile(tokenCachePath)
	ts, err := m.DeviceAuthorizationGrant(ctx, tenantID, clientID, scopes, nil)
	if err != nil {
		log.Fatal(err)
	}
	m.SaveFile(tokenCachePath)

	httpClient := oauth2.NewClient(ctx, ts)
	...
```

### Client credentials grant

- [OAuth 2.0 client credentials grant flow]

```go
const (
	tenantID     = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	clientID     = "YYYYYYYY-YYYY-YYYY-YYYY-YYYYYYYYYYYY"
	clientSecret = "ZZZZZZZZZZZZZZZZZZZZZZZZ"
)

var scopes = []string{msauth.DefaultMSGraphScope}

	ctx := context.Background()
	m := msauth.NewManager()
	ts, err := m.ClientCredentialsGrant(ctx, tenantID, clientID, clientSecret, scopes)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := oauth2.NewClient(ctx, ts)
    ...
```

### Resource owner password credentials grant

- [OAuth 2.0 resource owner passowrd credentials grant flow]

```go
const (
	tenantID     = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	clientID     = "YYYYYYYY-YYYY-YYYY-YYYY-YYYYYYYYYYYY"
	clientSecret = "ZZZZZZZZZZZZZZZZZZZZZZZZ"
	username     = "user.name@your-domain.com"
	password     = "secure-password"
)

var scopes = []string{msauth.DefaultMSGraphScope}

	ctx := context.Background()
	m := msauth.NewManager()
	ts, err := m.ResourceOwnerPasswordGrant(ctx, tenantID, clientID, clientSecret, username, password, scopes)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := oauth2.NewClient(ctx, ts)
    ...
```

### Authorization code grant

- [OAuth 2.0 authorization code grant flow]
- Not yet implemented.

[Microsoft identity platform]: https://docs.microsoft.com/en-us/azure/active-directory/develop/
[v2.0 endpoint]: https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-overview
[OAuth 2.0 device authorization grant flow]: https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-device-code
[OAuth 2.0 client credentials grant flow]: https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-client-creds-grant-flow
[OAuth 2.0 authorization code grant flow]: https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-auth-code-flow
[OAuth 2.0 resource owner passowrd credentials grant flow]: https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth-ropc
