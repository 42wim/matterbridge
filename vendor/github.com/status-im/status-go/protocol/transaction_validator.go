package protocol

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"time"

	"math/big"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	coretypes "github.com/status-im/status-go/eth-node/core/types"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/common"
)

const (
	transferFunction        = "a9059cbb"
	tokenTransferDataLength = 68
	transactionHashLength   = 66
)

type TransactionValidator struct {
	persistence *sqlitePersistence
	addresses   map[string]bool
	client      EthClient
	logger      *zap.Logger
}

var invalidResponse = &VerifyTransactionResponse{Valid: false}

type TransactionToValidate struct {
	TransactionHash string
	CommandID       string
	MessageID       string
	RetryCount      int
	// First seen indicates the whisper timestamp of the first time we seen this
	FirstSeen uint64
	// Validate indicates whether we should be validating this transaction
	Validate  bool
	Signature []byte
	From      *ecdsa.PublicKey
}

func NewTransactionValidator(addresses []types.Address, persistence *sqlitePersistence, client EthClient, logger *zap.Logger) *TransactionValidator {
	addressesMap := make(map[string]bool)
	for _, a := range addresses {
		addressesMap[strings.ToLower(a.Hex())] = true
	}
	logger.Debug("Checking addresses", zap.Any("addrse", addressesMap))

	return &TransactionValidator{
		persistence: persistence,
		addresses:   addressesMap,
		logger:      logger,
		client:      client,
	}
}

type EthClient interface {
	TransactionByHash(context.Context, types.Hash) (coretypes.Message, coretypes.TransactionStatus, error)
}

func (t *TransactionValidator) verifyTransactionSignature(ctx context.Context, from *ecdsa.PublicKey, address types.Address, transactionHash string, signature []byte) error {
	publicKeyBytes := crypto.FromECDSAPub(from)

	if len(transactionHash) != transactionHashLength {
		return errors.New("wrong transaction hash length")
	}

	hashBytes, err := hex.DecodeString(transactionHash[2:])
	if err != nil {
		return err
	}
	signatureMaterial := append(publicKeyBytes, hashBytes...)

	// We take a copy as EcRecover modifies the byte slice
	signatureCopy := make([]byte, len(signature))
	copy(signatureCopy, signature)
	extractedAddress, err := crypto.EcRecover(ctx, signatureMaterial, signatureCopy)
	if err != nil {
		return err
	}

	if extractedAddress != address {
		return errors.New("failed to verify signature")
	}
	return nil
}

func (t *TransactionValidator) validateTokenTransfer(parameters *common.CommandParameters, transaction coretypes.Message) (*VerifyTransactionResponse, error) {

	data := transaction.Data()
	if len(data) != tokenTransferDataLength {
		return nil, errors.New(fmt.Sprintf("wrong data length: %d", len(data)))
	}

	functionCalled := hex.EncodeToString(data[:4])

	if functionCalled != transferFunction {
		return invalidResponse, nil
	}

	actualContractAddress := strings.ToLower(transaction.To().Hex())

	if parameters.Contract != "" && actualContractAddress != parameters.Contract {
		return invalidResponse, nil
	}

	to := types.EncodeHex(data[16:36])

	if !t.validateToAddress(parameters.Address, to) {
		return invalidResponse, nil
	}

	value := data[36:]
	amount := new(big.Int).SetBytes(value)

	if parameters.Value != "" {
		advertisedAmount, ok := new(big.Int).SetString(parameters.Value, 10)
		if !ok {
			return nil, errors.New("can't parse amount")
		}

		return &VerifyTransactionResponse{
			Value:           parameters.Value,
			Contract:        actualContractAddress,
			Address:         to,
			AccordingToSpec: amount.Cmp(advertisedAmount) == 0,
			Valid:           true,
		}, nil
	}

	return &VerifyTransactionResponse{
		Value:           amount.String(),
		Address:         to,
		Contract:        actualContractAddress,
		AccordingToSpec: false,
		Valid:           true,
	}, nil

}

func (t *TransactionValidator) validateToAddress(specifiedTo, actualTo string) bool {
	if len(specifiedTo) != 0 && (!strings.EqualFold(specifiedTo, actualTo) || !t.addresses[strings.ToLower(actualTo)]) {
		return false
	}

	return t.addresses[actualTo]
}

func (t *TransactionValidator) validateEthereumTransfer(parameters *common.CommandParameters, transaction coretypes.Message) (*VerifyTransactionResponse, error) {
	toAddress := strings.ToLower(transaction.To().Hex())

	if !t.validateToAddress(parameters.Address, toAddress) {
		return invalidResponse, nil
	}

	amount := transaction.Value()
	if parameters.Value != "" {
		advertisedAmount, ok := new(big.Int).SetString(parameters.Value, 10)
		if !ok {
			return nil, errors.New("can't parse amount")
		}
		return &VerifyTransactionResponse{
			AccordingToSpec: amount.Cmp(advertisedAmount) == 0,
			Valid:           true,
			Value:           amount.String(),
			Address:         toAddress,
		}, nil
	}

	return &VerifyTransactionResponse{
		AccordingToSpec: false,
		Valid:           true,
		Value:           amount.String(),
		Address:         toAddress,
	}, nil
}

type VerifyTransactionResponse struct {
	Pending bool
	// AccordingToSpec means that the transaction is valid,
	// the user should be notified, but is not the same as
	// what was requested, for example because the value is different
	AccordingToSpec bool
	// Valid means that the transaction is valid
	Valid bool
	// The actual value received
	Value string
	// The contract used in case of tokens
	Contract string
	// The address the transaction was actually sent
	Address string

	Message     *common.Message
	Transaction *TransactionToValidate
}

// validateTransaction validates a transaction and returns a response.
// If a negative response is returned, i.e `Valid` is false, it should
// not be retried.
// If an error is returned, validation can be retried.
func (t *TransactionValidator) validateTransaction(ctx context.Context, message coretypes.Message, parameters *common.CommandParameters, from *ecdsa.PublicKey) (*VerifyTransactionResponse, error) {
	fromAddress := types.BytesToAddress(message.From().Bytes())

	err := t.verifyTransactionSignature(ctx, from, fromAddress, parameters.TransactionHash, parameters.Signature)
	if err != nil {
		t.logger.Error("failed validating signature", zap.Error(err))
		return invalidResponse, nil
	}

	if len(message.Data()) != 0 {
		t.logger.Debug("Validating token")
		return t.validateTokenTransfer(parameters, message)
	}

	t.logger.Debug("Validating eth")
	return t.validateEthereumTransfer(parameters, message)
}

func (t *TransactionValidator) ValidateTransactions(ctx context.Context) ([]*VerifyTransactionResponse, error) {
	if t.client == nil {
		return nil, nil
	}
	var response []*VerifyTransactionResponse
	t.logger.Debug("Started validating transactions")
	transactions, err := t.persistence.TransactionsToValidate()
	if err != nil {
		return nil, err
	}

	t.logger.Debug("Transactions to validated", zap.Any("transactions", transactions))

	for _, transaction := range transactions {
		var validationResult *VerifyTransactionResponse
		t.logger.Debug("Validating transaction", zap.Any("transaction", transaction))
		if transaction.CommandID != "" {
			chatID := contactIDFromPublicKey(transaction.From)
			message, err := t.persistence.MessageByCommandID(chatID, transaction.CommandID)
			if err != nil {
				t.logger.Error("error pulling message", zap.Error(err))
				return nil, err
			}
			if message == nil {
				t.logger.Info("No message found, ignoring transaction")
				// This is not a valid case, ignore transaction
				transaction.Validate = false
				transaction.RetryCount++
				err = t.persistence.UpdateTransactionToValidate(transaction)
				if err != nil {
					return nil, err
				}
				continue

			}
			commandParameters := message.CommandParameters
			commandParameters.TransactionHash = transaction.TransactionHash
			commandParameters.Signature = transaction.Signature
			validationResult, err = t.ValidateTransaction(ctx, message.CommandParameters, transaction.From)
			if err != nil {
				t.logger.Error("Error validating transaction", zap.Error(err))
				continue
			}
			validationResult.Message = message
		} else {
			commandParameters := &common.CommandParameters{}
			commandParameters.TransactionHash = transaction.TransactionHash
			commandParameters.Signature = transaction.Signature

			validationResult, err = t.ValidateTransaction(ctx, commandParameters, transaction.From)
			if err != nil {
				t.logger.Error("Error validating transaction", zap.Error(err))
				continue
			}
		}

		if validationResult.Pending {
			t.logger.Debug("Pending transaction skipping")
			// Check if we should stop updating
			continue
		}

		// Mark transaction as valid
		transaction.Validate = false
		transaction.RetryCount++
		err = t.persistence.UpdateTransactionToValidate(transaction)
		if err != nil {
			return nil, err
		}

		if !validationResult.Valid {
			t.logger.Debug("Transaction not valid")
			continue
		}
		t.logger.Debug("Transaction valid")
		validationResult.Transaction = transaction
		response = append(response, validationResult)
	}
	return response, nil
}

func (t *TransactionValidator) ValidateTransaction(ctx context.Context, parameters *common.CommandParameters, from *ecdsa.PublicKey) (*VerifyTransactionResponse, error) {
	t.logger.Debug("validating transaction", zap.Any("transaction", parameters), zap.Any("from", from))
	hash := parameters.TransactionHash
	c, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	message, status, err := t.client.TransactionByHash(c, types.HexToHash(hash))
	if err != nil {
		return nil, err
	}
	switch status {
	case coretypes.TransactionStatusPending:
		t.logger.Debug("Transaction pending")
		return &VerifyTransactionResponse{Pending: true}, nil
	case coretypes.TransactionStatusFailed:

		return invalidResponse, nil
	}

	return t.validateTransaction(ctx, message, parameters, from)
}
