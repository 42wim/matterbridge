package realtime

import (
	"github.com/Jeffail/gabs"
	"github.com/matterbridge/Rocket.Chat.Go.SDK/models"
)

// GetPermissions gets permissions
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/get-permissions
func (c *Client) GetPermissions() ([]models.Permission, error) {
	rawResponse, err := c.ddp.Call("permissions/get")
	if err != nil {
		return nil, err
	}

	document, _ := gabs.Consume(rawResponse)

	perms, _ := document.Children()

	var permissions []models.Permission

	for _, permission := range perms {
		var roles []string
		for _, role := range permission.Path("roles").Data().([]interface{}) {
			roles = append(roles, role.(string))
		}

		permissions = append(permissions, models.Permission{
			ID:    stringOrZero(permission.Path("_id").Data()),
			Roles: roles,
		})
	}

	return permissions, nil
}

// GetUserRoles gets current users roles
//
// https://rocket.chat/docs/developer-guides/realtime-api/method-calls/get-user-roles
func (c *Client) GetUserRoles() error {
	rawResponse, err := c.ddp.Call("getUserRoles")
	if err != nil {
		return err
	}

	document, _ := gabs.Consume(rawResponse)

	_, err = document.Children()
	// TODO: Figure out if this function is even useful if so return it
	//log.Println(roles)

	return nil
}
