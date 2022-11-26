// Auto-generated to Go types using avdl-compiler v1.4.10 (https://github.com/keybase/node-avdl-compiler)
//   Input file: ../client/protocol/avdl/keybase1/wot.avdl

package keybase1

import (
	"fmt"
)

type UsernameVerificationType string

func (o UsernameVerificationType) DeepCopy() UsernameVerificationType {
	return o
}

type WotProof struct {
	ProofType ProofType `codec:"proofType" json:"proof_type"`
	Name      string    `codec:"name" json:"name,omitempty"`
	Username  string    `codec:"username" json:"username,omitempty"`
	Protocol  string    `codec:"protocol" json:"protocol,omitempty"`
	Hostname  string    `codec:"hostname" json:"hostname,omitempty"`
	Domain    string    `codec:"domain" json:"domain,omitempty"`
}

func (o WotProof) DeepCopy() WotProof {
	return WotProof{
		ProofType: o.ProofType.DeepCopy(),
		Name:      o.Name,
		Username:  o.Username,
		Protocol:  o.Protocol,
		Hostname:  o.Hostname,
		Domain:    o.Domain,
	}
}

type WotProofUI struct {
	Type             string       `codec:"type" json:"type"`
	Value            string       `codec:"value" json:"value"`
	SiteIcon         []SizedImage `codec:"siteIcon" json:"siteIcon"`
	SiteIconDarkmode []SizedImage `codec:"siteIconDarkmode" json:"siteIconDarkmode"`
}

func (o WotProofUI) DeepCopy() WotProofUI {
	return WotProofUI{
		Type:  o.Type,
		Value: o.Value,
		SiteIcon: (func(x []SizedImage) []SizedImage {
			if x == nil {
				return nil
			}
			ret := make([]SizedImage, len(x))
			for i, v := range x {
				vCopy := v.DeepCopy()
				ret[i] = vCopy
			}
			return ret
		})(o.SiteIcon),
		SiteIconDarkmode: (func(x []SizedImage) []SizedImage {
			if x == nil {
				return nil
			}
			ret := make([]SizedImage, len(x))
			for i, v := range x {
				vCopy := v.DeepCopy()
				ret[i] = vCopy
			}
			return ret
		})(o.SiteIconDarkmode),
	}
}

type Confidence struct {
	UsernameVerifiedVia UsernameVerificationType `codec:"usernameVerifiedVia" json:"username_verified_via,omitempty"`
	Proofs              []WotProof               `codec:"proofs" json:"proofs,omitempty"`
	Other               string                   `codec:"other" json:"other,omitempty"`
}

func (o Confidence) DeepCopy() Confidence {
	return Confidence{
		UsernameVerifiedVia: o.UsernameVerifiedVia.DeepCopy(),
		Proofs: (func(x []WotProof) []WotProof {
			if x == nil {
				return nil
			}
			ret := make([]WotProof, len(x))
			for i, v := range x {
				vCopy := v.DeepCopy()
				ret[i] = vCopy
			}
			return ret
		})(o.Proofs),
		Other: o.Other,
	}
}

type WotReactionType int

const (
	WotReactionType_REJECT WotReactionType = 0
	WotReactionType_ACCEPT WotReactionType = 1
)

func (o WotReactionType) DeepCopy() WotReactionType { return o }

var WotReactionTypeMap = map[string]WotReactionType{
	"REJECT": 0,
	"ACCEPT": 1,
}

var WotReactionTypeRevMap = map[WotReactionType]string{
	0: "REJECT",
	1: "ACCEPT",
}

func (e WotReactionType) String() string {
	if v, ok := WotReactionTypeRevMap[e]; ok {
		return v
	}
	return fmt.Sprintf("%v", int(e))
}

type WotVouch struct {
	Status          WotStatusType `codec:"status" json:"status"`
	VouchProof      SigID         `codec:"vouchProof" json:"vouchProof"`
	Vouchee         UserVersion   `codec:"vouchee" json:"vouchee"`
	VoucheeUsername string        `codec:"voucheeUsername" json:"voucheeUsername"`
	Voucher         UserVersion   `codec:"voucher" json:"voucher"`
	VoucherUsername string        `codec:"voucherUsername" json:"voucherUsername"`
	VouchText       string        `codec:"vouchText" json:"vouchText"`
	VouchedAt       Time          `codec:"vouchedAt" json:"vouchedAt"`
	Confidence      Confidence    `codec:"confidence" json:"confidence"`
	Proofs          []WotProofUI  `codec:"proofs" json:"proofs"`
}

func (o WotVouch) DeepCopy() WotVouch {
	return WotVouch{
		Status:          o.Status.DeepCopy(),
		VouchProof:      o.VouchProof.DeepCopy(),
		Vouchee:         o.Vouchee.DeepCopy(),
		VoucheeUsername: o.VoucheeUsername,
		Voucher:         o.Voucher.DeepCopy(),
		VoucherUsername: o.VoucherUsername,
		VouchText:       o.VouchText,
		VouchedAt:       o.VouchedAt.DeepCopy(),
		Confidence:      o.Confidence.DeepCopy(),
		Proofs: (func(x []WotProofUI) []WotProofUI {
			if x == nil {
				return nil
			}
			ret := make([]WotProofUI, len(x))
			for i, v := range x {
				vCopy := v.DeepCopy()
				ret[i] = vCopy
			}
			return ret
		})(o.Proofs),
	}
}
