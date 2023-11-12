package walletconnect

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

const (
	SupportedEip155Namespace = "eip155"

	ProposeUserPairEvent = walletevent.EventType("WalletConnectProposeUserPair")
)

var (
	ErrorInvalidSessionProposal = errors.New("invalid session proposal")
	ErrorNamespaceNotSupported  = errors.New("namespace not supported")
	ErrorChainsNotSupported     = errors.New("chains not supported")
	ErrorInvalidParamsCount     = errors.New("invalid params count")
	ErrorInvalidAddressMsgIndex = errors.New("invalid address and/or msg index (must be 0 or 1)")
	ErrorMethodNotSupported     = errors.New("method not supported")
)

type Topic string

type Namespace struct {
	Methods  []string `json:"methods"`
	Chains   []string `json:"chains"` // CAIP-2 format e.g. ["eip155:1"]
	Events   []string `json:"events"`
	Accounts []string `json:"accounts,omitempty"` // CAIP-10 format e.g. ["eip155:1:0x453...228"]
}

type Metadata struct {
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Icons       []string `json:"icons"`
	Name        string   `json:"name"`
	VerifyURL   string   `json:"verifyUrl"`
}

type Proposer struct {
	PublicKey string   `json:"publicKey"`
	Metadata  Metadata `json:"metadata"`
}

type Verified struct {
	VerifyURL  string `json:"verifyUrl"`
	Validation string `json:"validation"`
	Origin     string `json:"origin"`
	IsScam     bool   `json:"isScam,omitempty"`
}

type VerifyContext struct {
	Verified Verified `json:"verified"`
}

type Params struct {
	ID                 int64                `json:"id"`
	PairingTopic       Topic                `json:"pairingTopic"`
	Expiry             int64                `json:"expiry"`
	RequiredNamespaces map[string]Namespace `json:"requiredNamespaces"`
	OptionalNamespaces map[string]Namespace `json:"optionalNamespaces"`
	Proposer           Proposer             `json:"proposer"`
	Verify             VerifyContext        `json:"verifyContext"`
}

type SessionProposal struct {
	ID     int64  `json:"id"`
	Params Params `json:"params"`
}

type PairSessionResponse struct {
	SessionProposal     SessionProposal      `json:"sessionProposal"`
	SupportedNamespaces map[string]Namespace `json:"supportedNamespaces"`
}

type RequestParams struct {
	Request struct {
		Method string            `json:"method"`
		Params []json.RawMessage `json:"params"`
	} `json:"request"`
	ChainID string `json:"chainId"`
}

type SessionRequest struct {
	ID     int64         `json:"id"`
	Topic  Topic         `json:"topic"`
	Params RequestParams `json:"params"`
	Verify VerifyContext `json:"verifyContext"`
}

type SessionDelete struct {
	ID    int64 `json:"id"`
	Topic Topic `json:"topic"`
}

type Session struct {
	Acknowledged       bool                 `json:"acknowledged"`
	Controller         string               `json:"controller"`
	Expiry             int64                `json:"expiry"`
	Namespaces         map[string]Namespace `json:"namespaces"`
	OptionalNamespaces map[string]Namespace `json:"optionalNamespaces"`
	PairingTopic       Topic                `json:"pairingTopic"`
	Peer               Proposer             `json:"peer"`
	Relay              json.RawMessage      `json:"relay"`
	RequiredNamespaces map[string]Namespace `json:"requiredNamespaces"`
	Self               Proposer             `json:"self"`
	Topic              Topic                `json:"topic"`
}

// Valid namespace
func (n *Namespace) Valid(namespaceName string, chainID *uint64) bool {
	if chainID == nil {
		if len(n.Chains) == 0 {
			log.Warn("namespace doesn't refer to any chain")
			return false
		}
		for _, caip2Str := range n.Chains {
			resolvedNamespaceName, _, err := parseCaip2ChainID(caip2Str)
			if err != nil {
				log.Warn("namespace chain not in caip2 format", "chain", caip2Str, "error", err)
				return false
			}

			if resolvedNamespaceName != namespaceName {
				log.Warn("namespace name doesn't match", "namespace", namespaceName, "chain", caip2Str)
				return false
			}
		}
	}
	return true
}

// Valid params
func (p *Params) Valid() bool {
	for key, ns := range p.RequiredNamespaces {
		var chainID *uint64
		if strings.Contains(key, ":") {
			resolvedNamespaceName, cID, err := parseCaip2ChainID(key)
			if err != nil {
				log.Warn("params validation failed CAIP-2", "str", key, "error", err)
				return false
			}
			key = resolvedNamespaceName
			chainID = &cID
		}

		if !isValidNamespaceName(key) {
			log.Warn("invalid namespace name", "namespace", key)
			return false
		}

		if !ns.Valid(key, chainID) {
			return false
		}
	}

	return true
}

// Valid session propsal
// https://specs.walletconnect.com/2.0/specs/clients/sign/namespaces#controller-side-validation-of-incoming-proposal-namespaces-wallet
func (p *SessionProposal) Valid() bool {
	return p.Params.Valid()
}

func sessionProposalToSupportedChain(caipChains []string, supportsChain func(uint64) bool) (chains []uint64, eipChains []string) {
	chains = make([]uint64, 0, 1)
	eipChains = make([]string, 0, 1)
	for _, caip2Str := range caipChains {
		_, chainID, err := parseCaip2ChainID(caip2Str)
		if err != nil {
			log.Warn("Failed parsing CAIP-2", "str", caip2Str, "error", err)
			continue
		}

		if !supportsChain(chainID) {
			continue
		}

		eipChains = append(eipChains, caip2Str)
		chains = append(chains, chainID)
	}
	return
}

func caip10Accounts(accounts []*accounts.Account, chains []uint64) []string {
	addresses := make([]string, 0, len(accounts)*len(chains))
	for _, acc := range accounts {
		for _, chainID := range chains {
			addresses = append(addresses, fmt.Sprintf("%s:%s:%s", SupportedEip155Namespace, strconv.FormatUint(chainID, 10), acc.Address.Hex()))
		}
	}
	return addresses
}
