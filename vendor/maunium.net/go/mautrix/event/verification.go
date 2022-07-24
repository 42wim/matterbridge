// Copyright (c) 2020 Nikos Filippakis
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"maunium.net/go/mautrix/id"
)

type VerificationMethod string

const VerificationMethodSAS VerificationMethod = "m.sas.v1"

// VerificationRequestEventContent represents the content of a m.key.verification.request to_device event.
// https://spec.matrix.org/v1.2/client-server-api/#mkeyverificationrequest
type VerificationRequestEventContent struct {
	// The device ID which is initiating the request.
	FromDevice id.DeviceID `json:"from_device"`
	// An opaque identifier for the verification request. Must be unique with respect to the devices involved.
	TransactionID string `json:"transaction_id,omitempty"`
	// The verification methods supported by the sender.
	Methods []VerificationMethod `json:"methods"`
	// The POSIX timestamp in milliseconds for when the request was made.
	Timestamp int64 `json:"timestamp,omitempty"`
	// The user that the event is sent to for in-room verification.
	To id.UserID `json:"to,omitempty"`
	// Original event ID for in-room verification.
	RelatesTo *RelatesTo `json:"m.relates_to,omitempty"`
}

func (vrec *VerificationRequestEventContent) SupportsVerificationMethod(meth VerificationMethod) bool {
	for _, supportedMeth := range vrec.Methods {
		if supportedMeth == meth {
			return true
		}
	}
	return false
}

type KeyAgreementProtocol string

const (
	KeyAgreementCurve25519           KeyAgreementProtocol = "curve25519"
	KeyAgreementCurve25519HKDFSHA256 KeyAgreementProtocol = "curve25519-hkdf-sha256"
)

type VerificationHashMethod string

const VerificationHashSHA256 VerificationHashMethod = "sha256"

type MACMethod string

const HKDFHMACSHA256 MACMethod = "hkdf-hmac-sha256"

type SASMethod string

const (
	SASDecimal SASMethod = "decimal"
	SASEmoji   SASMethod = "emoji"
)

// VerificationStartEventContent represents the content of a m.key.verification.start to_device event.
// https://spec.matrix.org/v1.2/client-server-api/#mkeyverificationstartmsasv1
type VerificationStartEventContent struct {
	// The device ID which is initiating the process.
	FromDevice id.DeviceID `json:"from_device"`
	// An opaque identifier for the verification process. Must be unique with respect to the devices involved.
	TransactionID string `json:"transaction_id,omitempty"`
	// The verification method to use.
	Method VerificationMethod `json:"method"`
	// The key agreement protocols the sending device understands.
	KeyAgreementProtocols []KeyAgreementProtocol `json:"key_agreement_protocols"`
	// The hash methods the sending device understands.
	Hashes []VerificationHashMethod `json:"hashes"`
	// The message authentication codes that the sending device understands.
	MessageAuthenticationCodes []MACMethod `json:"message_authentication_codes"`
	// The SAS methods the sending device (and the sending device's user) understands.
	ShortAuthenticationString []SASMethod `json:"short_authentication_string"`
	// The user that the event is sent to for in-room verification.
	To id.UserID `json:"to,omitempty"`
	// Original event ID for in-room verification.
	RelatesTo *RelatesTo `json:"m.relates_to,omitempty"`
}

func (vsec *VerificationStartEventContent) SupportsKeyAgreementProtocol(proto KeyAgreementProtocol) bool {
	for _, supportedProto := range vsec.KeyAgreementProtocols {
		if supportedProto == proto {
			return true
		}
	}
	return false
}

func (vsec *VerificationStartEventContent) SupportsHashMethod(alg VerificationHashMethod) bool {
	for _, supportedAlg := range vsec.Hashes {
		if supportedAlg == alg {
			return true
		}
	}
	return false
}

func (vsec *VerificationStartEventContent) SupportsMACMethod(meth MACMethod) bool {
	for _, supportedMeth := range vsec.MessageAuthenticationCodes {
		if supportedMeth == meth {
			return true
		}
	}
	return false
}

func (vsec *VerificationStartEventContent) SupportsSASMethod(meth SASMethod) bool {
	for _, supportedMeth := range vsec.ShortAuthenticationString {
		if supportedMeth == meth {
			return true
		}
	}
	return false
}

func (vsec *VerificationStartEventContent) GetRelatesTo() *RelatesTo {
	if vsec.RelatesTo == nil {
		vsec.RelatesTo = &RelatesTo{}
	}
	return vsec.RelatesTo
}

func (vsec *VerificationStartEventContent) OptionalGetRelatesTo() *RelatesTo {
	return vsec.RelatesTo
}

func (vsec *VerificationStartEventContent) SetRelatesTo(rel *RelatesTo) {
	vsec.RelatesTo = rel
}

// VerificationReadyEventContent represents the content of a m.key.verification.ready event.
// https://spec.matrix.org/v1.2/client-server-api/#mkeyverificationready
type VerificationReadyEventContent struct {
	// The device ID which accepted the process.
	FromDevice id.DeviceID `json:"from_device"`
	// The verification methods supported by the sender.
	Methods []VerificationMethod `json:"methods"`
	// Original event ID for in-room verification.
	RelatesTo *RelatesTo `json:"m.relates_to,omitempty"`
}

var _ Relatable = (*VerificationReadyEventContent)(nil)

func (vrec *VerificationReadyEventContent) GetRelatesTo() *RelatesTo {
	if vrec.RelatesTo == nil {
		vrec.RelatesTo = &RelatesTo{}
	}
	return vrec.RelatesTo
}

func (vrec *VerificationReadyEventContent) OptionalGetRelatesTo() *RelatesTo {
	return vrec.RelatesTo
}

func (vrec *VerificationReadyEventContent) SetRelatesTo(rel *RelatesTo) {
	vrec.RelatesTo = rel
}

// VerificationAcceptEventContent represents the content of a m.key.verification.accept to_device event.
// https://spec.matrix.org/v1.2/client-server-api/#mkeyverificationaccept
type VerificationAcceptEventContent struct {
	// An opaque identifier for the verification process. Must be the same as the one used for the m.key.verification.start message.
	TransactionID string `json:"transaction_id,omitempty"`
	// The verification method to use.
	Method VerificationMethod `json:"method"`
	// The key agreement protocol the device is choosing to use, out of the options in the m.key.verification.start message.
	KeyAgreementProtocol KeyAgreementProtocol `json:"key_agreement_protocol"`
	// The hash method the device is choosing to use, out of the options in the m.key.verification.start message.
	Hash VerificationHashMethod `json:"hash"`
	// The message authentication code the device is choosing to use, out of the options in the m.key.verification.start message.
	MessageAuthenticationCode MACMethod `json:"message_authentication_code"`
	// The SAS methods both devices involved in the verification process understand. Must be a subset of the options in the m.key.verification.start message.
	ShortAuthenticationString []SASMethod `json:"short_authentication_string"`
	// The hash (encoded as unpadded base64) of the concatenation of the device's ephemeral public key (encoded as unpadded base64) and the canonical JSON representation of the m.key.verification.start message.
	Commitment string `json:"commitment"`
	// The user that the event is sent to for in-room verification.
	To id.UserID `json:"to,omitempty"`
	// Original event ID for in-room verification.
	RelatesTo *RelatesTo `json:"m.relates_to,omitempty"`
}

func (vaec *VerificationAcceptEventContent) GetRelatesTo() *RelatesTo {
	if vaec.RelatesTo == nil {
		vaec.RelatesTo = &RelatesTo{}
	}
	return vaec.RelatesTo
}

func (vaec *VerificationAcceptEventContent) OptionalGetRelatesTo() *RelatesTo {
	return vaec.RelatesTo
}

func (vaec *VerificationAcceptEventContent) SetRelatesTo(rel *RelatesTo) {
	vaec.RelatesTo = rel
}

// VerificationKeyEventContent represents the content of a m.key.verification.key to_device event.
// https://spec.matrix.org/v1.2/client-server-api/#mkeyverificationkey
type VerificationKeyEventContent struct {
	// An opaque identifier for the verification process. Must be the same as the one used for the m.key.verification.start message.
	TransactionID string `json:"transaction_id,omitempty"`
	// The device's ephemeral public key, encoded as unpadded base64.
	Key string `json:"key"`
	// The user that the event is sent to for in-room verification.
	To id.UserID `json:"to,omitempty"`
	// Original event ID for in-room verification.
	RelatesTo *RelatesTo `json:"m.relates_to,omitempty"`
}

func (vkec *VerificationKeyEventContent) GetRelatesTo() *RelatesTo {
	if vkec.RelatesTo == nil {
		vkec.RelatesTo = &RelatesTo{}
	}
	return vkec.RelatesTo
}

func (vkec *VerificationKeyEventContent) OptionalGetRelatesTo() *RelatesTo {
	return vkec.RelatesTo
}

func (vkec *VerificationKeyEventContent) SetRelatesTo(rel *RelatesTo) {
	vkec.RelatesTo = rel
}

// VerificationMacEventContent represents the content of a m.key.verification.mac to_device event.
// https://spec.matrix.org/v1.2/client-server-api/#mkeyverificationmac
type VerificationMacEventContent struct {
	// An opaque identifier for the verification process. Must be the same as the one used for the m.key.verification.start message.
	TransactionID string `json:"transaction_id,omitempty"`
	// A map of the key ID to the MAC of the key, using the algorithm in the verification process. The MAC is encoded as unpadded base64.
	Mac map[id.KeyID]string `json:"mac"`
	// The MAC of the comma-separated, sorted, list of key IDs given in the mac property, encoded as unpadded base64.
	Keys string `json:"keys"`
	// The user that the event is sent to for in-room verification.
	To id.UserID `json:"to,omitempty"`
	// Original event ID for in-room verification.
	RelatesTo *RelatesTo `json:"m.relates_to,omitempty"`
}

func (vmec *VerificationMacEventContent) GetRelatesTo() *RelatesTo {
	if vmec.RelatesTo == nil {
		vmec.RelatesTo = &RelatesTo{}
	}
	return vmec.RelatesTo
}

func (vmec *VerificationMacEventContent) OptionalGetRelatesTo() *RelatesTo {
	return vmec.RelatesTo
}

func (vmec *VerificationMacEventContent) SetRelatesTo(rel *RelatesTo) {
	vmec.RelatesTo = rel
}

type VerificationCancelCode string

const (
	VerificationCancelByUser             VerificationCancelCode = "m.user"
	VerificationCancelByTimeout          VerificationCancelCode = "m.timeout"
	VerificationCancelUnknownTransaction VerificationCancelCode = "m.unknown_transaction"
	VerificationCancelUnknownMethod      VerificationCancelCode = "m.unknown_method"
	VerificationCancelUnexpectedMessage  VerificationCancelCode = "m.unexpected_message"
	VerificationCancelKeyMismatch        VerificationCancelCode = "m.key_mismatch"
	VerificationCancelUserMismatch       VerificationCancelCode = "m.user_mismatch"
	VerificationCancelInvalidMessage     VerificationCancelCode = "m.invalid_message"
	VerificationCancelAccepted           VerificationCancelCode = "m.accepted"
	VerificationCancelSASMismatch        VerificationCancelCode = "m.mismatched_sas"
	VerificationCancelCommitmentMismatch VerificationCancelCode = "m.mismatched_commitment"
)

// VerificationCancelEventContent represents the content of a m.key.verification.cancel to_device event.
// https://spec.matrix.org/v1.2/client-server-api/#mkeyverificationcancel
type VerificationCancelEventContent struct {
	// The opaque identifier for the verification process/request.
	TransactionID string `json:"transaction_id,omitempty"`
	// A human readable description of the code. The client should only rely on this string if it does not understand the code.
	Reason string `json:"reason"`
	// The error code for why the process/request was cancelled by the user.
	Code VerificationCancelCode `json:"code"`
	// The user that the event is sent to for in-room verification.
	To id.UserID `json:"to,omitempty"`
	// Original event ID for in-room verification.
	RelatesTo *RelatesTo `json:"m.relates_to,omitempty"`
}

func (vcec *VerificationCancelEventContent) GetRelatesTo() *RelatesTo {
	if vcec.RelatesTo == nil {
		vcec.RelatesTo = &RelatesTo{}
	}
	return vcec.RelatesTo
}

func (vcec *VerificationCancelEventContent) OptionalGetRelatesTo() *RelatesTo {
	return vcec.RelatesTo
}

func (vcec *VerificationCancelEventContent) SetRelatesTo(rel *RelatesTo) {
	vcec.RelatesTo = rel
}
