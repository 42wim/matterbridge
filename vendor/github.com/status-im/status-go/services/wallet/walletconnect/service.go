package walletconnect

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"

	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc/network"
	"github.com/status-im/status-go/services/wallet/transfer"
)

type Service struct {
	db             *sql.DB
	networkManager *network.Manager
	accountsDB     *accounts.Database
	eventFeed      *event.Feed

	transactionManager *transfer.TransactionManager
	gethManager        *account.GethManager

	config *params.NodeConfig
}

func NewService(db *sql.DB, networkManager *network.Manager, accountsDB *accounts.Database,
	transactionManager *transfer.TransactionManager, gethManager *account.GethManager, eventFeed *event.Feed,
	config *params.NodeConfig) *Service {
	return &Service{
		db:                 db,
		networkManager:     networkManager,
		accountsDB:         accountsDB,
		eventFeed:          eventFeed,
		transactionManager: transactionManager,
		gethManager:        gethManager,
		config:             config,
	}
}

func (s *Service) PairSessionProposal(proposal SessionProposal) (*PairSessionResponse, error) {
	if !proposal.Valid() {
		return nil, ErrorInvalidSessionProposal
	}

	var (
		chains    []uint64
		eipChains []string
	)

	if len(proposal.Params.RequiredNamespaces) == 0 {
		// return all we support
		allChains, err := s.networkManager.GetAll()
		if err != nil {
			return nil, fmt.Errorf("failed to get all chains: %w", err)
		}
		for _, chain := range allChains {
			chains = append(chains, chain.ChainID)
			eipChains = append(eipChains, fmt.Sprintf("%s:%d", SupportedEip155Namespace, chain.ChainID))
		}
	} else {
		var proposedChains []string
		for key, ns := range proposal.Params.RequiredNamespaces {
			if !strings.Contains(key, SupportedEip155Namespace) {
				log.Warn("Some namespaces are not supported; wanted: ", key, "; supported: ", SupportedEip155Namespace)
				return nil, ErrorNamespaceNotSupported
			}

			if strings.Contains(key, ":") {
				proposedChains = append(proposedChains, key)
			} else {
				proposedChains = append(proposedChains, ns.Chains...)
			}
		}

		chains, eipChains = sessionProposalToSupportedChain(proposedChains, func(chainID uint64) bool {
			return s.networkManager.Find(chainID) != nil
		})

		if len(chains) != len(proposedChains) {
			log.Warn("Some chains are not supported; wanted: ", proposedChains, "; supported: ", chains)
			return nil, ErrorChainsNotSupported
		}
	}

	activeAccounts, err := s.accountsDB.GetActiveAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to get active accounts: %w", err)
	}

	allWalletAccountsReadyForTransaction := make([]*accounts.Account, 0, 1)
	for _, acc := range activeAccounts {
		if !acc.IsWalletAccountReadyForTransaction() {
			continue
		}
		allWalletAccountsReadyForTransaction = append(allWalletAccountsReadyForTransaction, acc)
	}

	result := &PairSessionResponse{
		SessionProposal: proposal,
		SupportedNamespaces: map[string]Namespace{
			SupportedEip155Namespace: Namespace{
				Methods: []string{params.SendTransactionMethodName,
					params.SendRawTransactionMethodName,
					params.PersonalSignMethodName,
					params.SignMethodName,
					params.SignTransactionMethodName,
					params.SignTypedDataMethodName,
					params.SignTypedDataV3MethodName,
					params.SignTypedDataV4MethodName,
					params.WalletSwitchEthereumChainMethodName,
				},
				Events:   []string{"accountsChanged", "chainChanged"},
				Chains:   eipChains,
				Accounts: caip10Accounts(allWalletAccountsReadyForTransaction, chains),
			},
		},
	}

	// TODO #12434: respond async
	return result, nil
}

func (s *Service) SaveOrUpdateSession(session Session) error {
	var icon string
	if len(session.Peer.Metadata.Icons) > 0 {
		icon = session.Peer.Metadata.Icons[0]
	}

	return UpsertSession(s.db, DbSession{
		Topic:           session.Topic,
		PairingTopic:    session.PairingTopic,
		Expiry:          session.Expiry,
		Active:          true,
		DappName:        session.Peer.Metadata.Name,
		DappURL:         session.Peer.Metadata.URL,
		DappDescription: session.Peer.Metadata.Description,
		DappIcon:        icon,
		DappVerifyURL:   session.Peer.Metadata.VerifyURL,
		DappPublicKey:   session.Peer.PublicKey,
	})
}

func (s *Service) ChangeSessionState(topic Topic, active bool) error {
	return ChangeSessionState(s.db, topic, active)
}

func (s *Service) SessionRequest(request SessionRequest) (response *transfer.TxResponse, err error) {
	// TODO #12434: should we check topic for validity? It might make sense if we
	// want to cache the paired sessions

	if request.Params.Request.Method == params.SendTransactionMethodName {
		return s.buildTransaction(request)
	} else if request.Params.Request.Method == params.SignTransactionMethodName {
		return s.buildTransaction(request)
	} else if request.Params.Request.Method == params.PersonalSignMethodName {
		return s.buildMessage(request, 1, 0, false)
	} else if request.Params.Request.Method == params.SignMethodName {
		return s.buildMessage(request, 0, 1, false)
	} else if request.Params.Request.Method == params.SignTypedDataMethodName ||
		request.Params.Request.Method == params.SignTypedDataV3MethodName ||
		request.Params.Request.Method == params.SignTypedDataV4MethodName {
		return s.buildMessage(request, 0, 1, true)
	}

	// TODO #12434: respond async
	return nil, ErrorMethodNotSupported
}

func (s *Service) AuthRequest(address common.Address, authMessage string) (*transfer.TxResponse, error) {
	account, err := s.accountsDB.GetAccountByAddress(types.Address(address))
	if err != nil {
		return nil, fmt.Errorf("failed to get active account: %w", err)
	}

	kp, err := s.accountsDB.GetKeypairByKeyUID(account.KeyUID)
	if err != nil {
		return nil, err
	}

	byteArray := []byte(authMessage)
	hash := crypto.TextHash(byteArray)

	return &transfer.TxResponse{
		KeyUID:        account.KeyUID,
		Address:       account.Address,
		AddressPath:   account.Path,
		SignOnKeycard: kp.MigratedToKeycard(),
		MessageToSign: types.HexBytes(hash),
	}, nil
}
