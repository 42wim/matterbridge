// Auto-generated to Go types using avdl-compiler v1.4.8 (https://github.com/keybase/node-avdl-compiler)
//   Input file: ../client/protocol/avdl/keybase1/invite_friends.avdl

package keybase1

type InviteCounts struct {
	InviteCount      int     `codec:"inviteCount" json:"inviteCount"`
	PercentageChange float64 `codec:"percentageChange" json:"percentageChange"`
	ShowNumInvites   bool    `codec:"showNumInvites" json:"showNumInvites"`
	ShowFire         bool    `codec:"showFire" json:"showFire"`
	TooltipMarkdown  string  `codec:"tooltipMarkdown" json:"tooltipMarkdown"`
}

func (o InviteCounts) DeepCopy() InviteCounts {
	return InviteCounts{
		InviteCount:      o.InviteCount,
		PercentageChange: o.PercentageChange,
		ShowNumInvites:   o.ShowNumInvites,
		ShowFire:         o.ShowFire,
		TooltipMarkdown:  o.TooltipMarkdown,
	}
}

type EmailInvites struct {
	CommaSeparatedEmailsFromUser *string         `codec:"commaSeparatedEmailsFromUser,omitempty" json:"commaSeparatedEmailsFromUser,omitempty"`
	EmailsFromContacts           *[]EmailAddress `codec:"emailsFromContacts,omitempty" json:"emailsFromContacts,omitempty"`
}

func (o EmailInvites) DeepCopy() EmailInvites {
	return EmailInvites{
		CommaSeparatedEmailsFromUser: (func(x *string) *string {
			if x == nil {
				return nil
			}
			tmp := (*x)
			return &tmp
		})(o.CommaSeparatedEmailsFromUser),
		EmailsFromContacts: (func(x *[]EmailAddress) *[]EmailAddress {
			if x == nil {
				return nil
			}
			tmp := (func(x []EmailAddress) []EmailAddress {
				if x == nil {
					return nil
				}
				ret := make([]EmailAddress, len(x))
				for i, v := range x {
					vCopy := v.DeepCopy()
					ret[i] = vCopy
				}
				return ret
			})((*x))
			return &tmp
		})(o.EmailsFromContacts),
	}
}
