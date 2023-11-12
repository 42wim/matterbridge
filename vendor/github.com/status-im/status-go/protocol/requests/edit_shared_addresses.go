package requests

import (
	"errors"

	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
)

var ErrInvalidCommunityID = errors.New("invalid community id")
var ErrMissingPassword = errors.New("password is necessary when sending a list of addresses")
var ErrMissingSharedAddresses = errors.New("list of shared addresses is needed")
var ErrMissingAirdropAddress = errors.New("airdropAddress is needed")
var ErrNoAirdropAddressAmongAddressesToReveal = errors.New("airdropAddress must be in the set of addresses to reveal")
var ErrInvalidSignature = errors.New("invalid signature")

type EditSharedAddresses struct {
	CommunityID       types.HexBytes   `json:"communityId"`
	AddressesToReveal []string         `json:"addressesToReveal"`
	Signatures        []types.HexBytes `json:"signatures"` // the order of signatures should match the order of addresses
	AirdropAddress    string           `json:"airdropAddress"`
}

func (j *EditSharedAddresses) Validate() error {
	if len(j.CommunityID) == 0 {
		return ErrInvalidCommunityID
	}

	if len(j.AddressesToReveal) == 0 {
		return ErrMissingSharedAddresses
	}

	if j.AirdropAddress == "" {
		return ErrMissingAirdropAddress
	}

	found := false
	for _, address := range j.AddressesToReveal {
		if address == j.AirdropAddress {
			found = true
			break
		}
	}

	if !found {
		return ErrNoAirdropAddressAmongAddressesToReveal
	}

	for _, signature := range j.Signatures {
		if len(signature) > 0 && len(signature) != crypto.SignatureLength {
			return ErrInvalidSignature
		}
	}

	return nil
}
