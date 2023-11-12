package static

import (
	"errors"

	"github.com/waku-org/go-zerokit-rln/rln"
)

func Setup(index rln.MembershipIndex) ([]rln.IDCommitment, rln.IdentityCredential, error) {
	// static group
	groupKeys := rln.STATIC_GROUP_KEYS
	groupSize := rln.STATIC_GROUP_SIZE

	// validate the user-supplied membership index
	if index >= rln.MembershipIndex(groupSize) {
		return nil, rln.IdentityCredential{}, errors.New("wrong membership index")
	}

	// create a sequence of MembershipKeyPairs from the group keys (group keys are in string format)
	credentials, err := rln.ToIdentityCredentials(groupKeys)
	if err != nil {
		return nil, rln.IdentityCredential{}, errors.New("invalid data on group keypairs")
	}

	// extract id commitment keys
	var groupOpt []rln.IDCommitment
	for _, c := range credentials {
		groupOpt = append(groupOpt, c.IDCommitment)
	}

	return groupOpt, credentials[index], nil
}
