package api // import "github.com/SevereCloud/vksdk/v2/api"

import (
	"github.com/SevereCloud/vksdk/v2/object"
)

// StorageGetResponse struct.
type StorageGetResponse []object.BaseRequestParam

// ToMap return map from StorageGetResponse.
func (s StorageGetResponse) ToMap() map[string]string {
	m := make(map[string]string)
	for _, item := range s {
		m[item.Key] = item.Value
	}

	return m
}

// StorageGet returns a value of variable with the name set by key parameter.
//
// StorageGet always return array!
//
// https://dev.vk.com/method/storage.get
func (vk *VK) StorageGet(params Params) (response StorageGetResponse, err error) {
	err = vk.RequestUnmarshal("storage.get", &response, params)

	return
}

// StorageGetKeysResponse struct.
type StorageGetKeysResponse []string

// StorageGetKeys returns the names of all variables.
//
// https://dev.vk.com/method/storage.getKeys
func (vk *VK) StorageGetKeys(params Params) (response StorageGetKeysResponse, err error) {
	err = vk.RequestUnmarshal("storage.getKeys", &response, params)
	return
}

// StorageSet saves a value of variable with the name set by key parameter.
//
// https://dev.vk.com/method/storage.set
func (vk *VK) StorageSet(params Params) (response int, err error) {
	err = vk.RequestUnmarshal("storage.set", &response, params)
	return
}
