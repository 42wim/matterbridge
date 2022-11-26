// Auto-generated to Go types using avdl-compiler v1.4.10 (https://github.com/keybase/node-avdl-compiler)
//   Input file: ../client/protocol/avdl/keybase1/teamsearch.avdl

package keybase1

type TeamSearchItem struct {
	Id          TeamID  `codec:"id" json:"id"`
	Name        string  `codec:"name" json:"name"`
	Description *string `codec:"description,omitempty" json:"description,omitempty"`
	MemberCount int     `codec:"memberCount" json:"memberCount"`
	LastActive  Time    `codec:"lastActive" json:"lastActive"`
	IsDemoted   bool    `codec:"isDemoted" json:"isDemoted"`
	InTeam      bool    `codec:"inTeam" json:"inTeam"`
}

func (o TeamSearchItem) DeepCopy() TeamSearchItem {
	return TeamSearchItem{
		Id:   o.Id.DeepCopy(),
		Name: o.Name,
		Description: (func(x *string) *string {
			if x == nil {
				return nil
			}
			tmp := (*x)
			return &tmp
		})(o.Description),
		MemberCount: o.MemberCount,
		LastActive:  o.LastActive.DeepCopy(),
		IsDemoted:   o.IsDemoted,
		InTeam:      o.InTeam,
	}
}

type TeamSearchExport struct {
	Items     map[TeamID]TeamSearchItem `codec:"items" json:"items"`
	Suggested []TeamID                  `codec:"suggested" json:"suggested"`
}

func (o TeamSearchExport) DeepCopy() TeamSearchExport {
	return TeamSearchExport{
		Items: (func(x map[TeamID]TeamSearchItem) map[TeamID]TeamSearchItem {
			if x == nil {
				return nil
			}
			ret := make(map[TeamID]TeamSearchItem, len(x))
			for k, v := range x {
				kCopy := k.DeepCopy()
				vCopy := v.DeepCopy()
				ret[kCopy] = vCopy
			}
			return ret
		})(o.Items),
		Suggested: (func(x []TeamID) []TeamID {
			if x == nil {
				return nil
			}
			ret := make([]TeamID, len(x))
			for i, v := range x {
				vCopy := v.DeepCopy()
				ret[i] = vCopy
			}
			return ret
		})(o.Suggested),
	}
}

type TeamSearchRes struct {
	Results []TeamSearchItem `codec:"results" json:"results"`
}

func (o TeamSearchRes) DeepCopy() TeamSearchRes {
	return TeamSearchRes{
		Results: (func(x []TeamSearchItem) []TeamSearchItem {
			if x == nil {
				return nil
			}
			ret := make([]TeamSearchItem, len(x))
			for i, v := range x {
				vCopy := v.DeepCopy()
				ret[i] = vCopy
			}
			return ret
		})(o.Results),
	}
}
