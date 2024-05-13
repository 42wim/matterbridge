package requests

type TogglePeerSyncingRequest struct {
	Enabled bool `json:"enabled"`
}

func (a *TogglePeerSyncingRequest) Validate() error {
	return nil
}
