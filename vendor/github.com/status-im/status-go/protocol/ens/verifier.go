package ens

import (
	"database/sql"
	"time"

	"go.uber.org/zap"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/eth-node/types"
	enstypes "github.com/status-im/status-go/eth-node/types/ens"
	"github.com/status-im/status-go/protocol/common"
)

type Verifier struct {
	node            types.Node
	online          bool
	persistence     *Persistence
	logger          *zap.Logger
	timesource      common.TimeSource
	subscriptions   []chan []*VerificationRecord
	rpcEndpoint     string
	contractAddress string
	quit            chan struct{}
}

func New(node types.Node, logger *zap.Logger, timesource common.TimeSource, db *sql.DB, rpcEndpoint, contractAddress string) *Verifier {
	persistence := NewPersistence(db)
	return &Verifier{
		node:            node,
		logger:          logger,
		persistence:     persistence,
		timesource:      timesource,
		rpcEndpoint:     rpcEndpoint,
		contractAddress: contractAddress,
		quit:            make(chan struct{}),
	}
}

func (v *Verifier) Start() error {
	go v.verifyLoop()
	return nil
}

func (v *Verifier) Stop() error {
	close(v.quit)

	return nil
}

// ENSVerified adds an already verified entry to the ens table
func (v *Verifier) ENSVerified(pk, ensName string, clock uint64) error {

	// Add returns nil if no record was available
	oldRecord, err := v.Add(pk, ensName, clock)
	if err != nil {
		return err
	}

	var record *VerificationRecord

	if oldRecord != nil {
		record = oldRecord
	} else {
		record = &VerificationRecord{PublicKey: pk, Name: ensName, Clock: clock}
	}

	record.VerifiedAt = clock
	record.Verified = true
	records := []*VerificationRecord{record}
	err = v.persistence.UpdateRecords(records)
	if err != nil {
		return err
	}
	v.publish(records)
	return nil
}

func (v *Verifier) GetVerifiedRecord(pk string) (*VerificationRecord, error) {
	return v.persistence.GetVerifiedRecord(pk)
}

func (v *Verifier) Add(pk, ensName string, clock uint64) (*VerificationRecord, error) {
	record := VerificationRecord{PublicKey: pk, Name: ensName, Clock: clock}
	return v.persistence.AddRecord(record)
}

func (v *Verifier) SetOnline(online bool) {
	v.online = online
}

func (v *Verifier) verifyLoop() {

	ticker := time.NewTicker(30 * time.Second)
	for {
		select {

		case <-v.quit:
			ticker.Stop()
			return
		case <-ticker.C:
			if !v.online || v.rpcEndpoint == "" || v.contractAddress == "" {
				continue
			}
			err := v.verify(v.rpcEndpoint, v.contractAddress)
			if err != nil {
				v.logger.Error("verify loop failed", zap.Error(err))
			}

		}
	}
}

func (v *Verifier) Subscribe() chan []*VerificationRecord {
	c := make(chan []*VerificationRecord)
	v.subscriptions = append(v.subscriptions, c)
	return c
}

func (v *Verifier) publish(records []*VerificationRecord) {
	v.logger.Info("publishing records", zap.Any("records", records))
	// Publish on channels, drop if buffer is full
	for _, s := range v.subscriptions {
		select {
		case s <- records:
		default:
			v.logger.Warn("ens subscription channel full, dropping message")
		}
	}

}

func (v *Verifier) ReverseResolve(address gethcommon.Address) (string, error) {
	verifier := v.node.NewENSVerifier(v.logger)
	return verifier.ReverseResolve(address, v.rpcEndpoint)
}

// Verify verifies that a registered ENS name matches the expected public key
func (v *Verifier) verify(rpcEndpoint, contractAddress string) error {
	v.logger.Debug("verifying ENS Names", zap.String("endpoint", rpcEndpoint))
	verifier := v.node.NewENSVerifier(v.logger)

	var ensDetails []enstypes.ENSDetails

	// Now in seconds
	now := v.timesource.GetCurrentTime() / 1000
	ensToBeVerified, err := v.persistence.GetENSToBeVerified(now)
	if err != nil {
		return err
	}

	recordsMap := make(map[string]*VerificationRecord)

	for _, r := range ensToBeVerified {
		recordsMap[r.PublicKey] = r
		ensDetails = append(ensDetails, enstypes.ENSDetails{
			PublicKeyString: r.PublicKey[2:],
			Name:            r.Name,
		})
		v.logger.Debug("verifying ens name", zap.Any("record", r))
	}

	ensResponse, err := verifier.CheckBatch(ensDetails, rpcEndpoint, contractAddress)
	if err != nil {
		v.logger.Error("failed to check batch", zap.Error(err))
		return err
	}

	var records []*VerificationRecord

	for _, details := range ensResponse {
		pk := "0x" + details.PublicKeyString
		record := recordsMap[pk]

		if details.Error == nil {
			record.Verified = details.Verified
			if !record.Verified {
				record.VerificationRetries++
			}
		} else {
			v.logger.Warn("Failed to resolve ens name",
				zap.String("name", details.Name),
				zap.String("publicKey", details.PublicKeyString),
				zap.Error(details.Error),
			)
			record.VerificationRetries++
		}
		record.VerifiedAt = now
		record.CalculateNextRetry()

		records = append(records, record)
	}

	err = v.persistence.UpdateRecords(records)
	if err != nil {

		v.logger.Error("failed to update records", zap.Error(err))
		return err
	}

	v.publish(records)

	return nil
}
