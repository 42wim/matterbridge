package requests

type BuildContact struct {
	PublicKey string `json:"publicKey"`
	ENSName   string `json:"ENSName"`
}
