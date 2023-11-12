package permissions

import (
	"context"
)

func NewAPI(db *Database) *API {
	return &API{db}
}

// API is class with methods available over RPC.
type API struct {
	db *Database
}

func (api *API) AddDappPermissions(ctx context.Context, perms DappPermissions) error {
	return api.db.AddPermissions(perms)
}

func (api *API) GetDappPermissions(ctx context.Context) ([]DappPermissions, error) {
	return api.db.GetPermissions()
}

func (api *API) DeleteDappPermissions(ctx context.Context, name string) error {
	return api.db.DeletePermission(name, "")
}

func (api *API) DeleteDappPermissionsByNameAndAddress(ctx context.Context, name string, address string) error {
	return api.db.DeletePermission(name, address)
}
