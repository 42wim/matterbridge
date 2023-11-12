package requests

type SetCustomNodes struct {
	CustomNodes map[string]string `json:"customNodes"`
}
