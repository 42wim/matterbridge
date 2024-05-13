package communities

import (
	"encoding/json"
	"sort"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/status-im/status-go/protocol/protobuf"
)

type CheckPermissionsResponse struct {
	Satisfied            bool                                      `json:"satisfied"`
	Permissions          map[string]*PermissionTokenCriteriaResult `json:"permissions"`
	ValidCombinations    []*AccountChainIDsCombination             `json:"validCombinations"`
	NetworksNotSupported bool                                      `json:"networksNotSupported"`
}

type CheckPermissionToJoinResponse = CheckPermissionsResponse

type HighestRoleResponse struct {
	Role      protobuf.CommunityTokenPermission_Type `json:"type"`
	Satisfied bool                                   `json:"satisfied"`
	Criteria  []*PermissionTokenCriteriaResult       `json:"criteria"`
}

var roleOrders = map[protobuf.CommunityTokenPermission_Type]int{
	protobuf.CommunityTokenPermission_BECOME_MEMBER:             1,
	protobuf.CommunityTokenPermission_CAN_VIEW_CHANNEL:          2,
	protobuf.CommunityTokenPermission_CAN_VIEW_AND_POST_CHANNEL: 3,
	protobuf.CommunityTokenPermission_BECOME_ADMIN:              4,
	protobuf.CommunityTokenPermission_BECOME_TOKEN_MASTER:       5,
	protobuf.CommunityTokenPermission_BECOME_TOKEN_OWNER:        6,
}

type ByRoleDesc []*HighestRoleResponse

func (a ByRoleDesc) Len() int      { return len(a) }
func (a ByRoleDesc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByRoleDesc) Less(i, j int) bool {
	return roleOrders[a[i].Role] > roleOrders[a[j].Role]
}

type rolesAndHighestRole struct {
	Roles       []*HighestRoleResponse
	HighestRole *HighestRoleResponse
}

func calculateRolesAndHighestRole(permissions map[string]*PermissionTokenCriteriaResult) *rolesAndHighestRole {
	item := &rolesAndHighestRole{}
	byRoleMap := make(map[protobuf.CommunityTokenPermission_Type]*HighestRoleResponse)
	for _, p := range permissions {
		if roleOrders[p.Role] == 0 {
			continue
		}
		if byRoleMap[p.Role] == nil {
			byRoleMap[p.Role] = &HighestRoleResponse{
				Role: p.Role,
			}
		}

		satisfied := true
		for _, tr := range p.TokenRequirements {
			if !tr.Satisfied {
				satisfied = false
				break
			}

		}

		if satisfied {
			byRoleMap[p.Role].Satisfied = true
			// we prepend
			byRoleMap[p.Role].Criteria = append([]*PermissionTokenCriteriaResult{p}, byRoleMap[p.Role].Criteria...)
		} else {
			// we append then
			byRoleMap[p.Role].Criteria = append(byRoleMap[p.Role].Criteria, p)
		}
	}
	if byRoleMap[protobuf.CommunityTokenPermission_BECOME_MEMBER] == nil {
		byRoleMap[protobuf.CommunityTokenPermission_BECOME_MEMBER] = &HighestRoleResponse{Satisfied: true, Role: protobuf.CommunityTokenPermission_BECOME_MEMBER}
	}
	for _, p := range byRoleMap {
		item.Roles = append(item.Roles, p)
	}

	sort.Sort(ByRoleDesc(item.Roles))
	for _, r := range item.Roles {
		if r.Satisfied {
			item.HighestRole = r
			break
		}

	}
	return item
}

func (c *CheckPermissionsResponse) MarshalJSON() ([]byte, error) {
	type CheckPermissionsTypeAlias struct {
		Satisfied            bool                                      `json:"satisfied"`
		Permissions          map[string]*PermissionTokenCriteriaResult `json:"permissions"`
		ValidCombinations    []*AccountChainIDsCombination             `json:"validCombinations"`
		Roles                []*HighestRoleResponse                    `json:"roles"`
		HighestRole          *HighestRoleResponse                      `json:"highestRole"`
		NetworksNotSupported bool                                      `json:"networksNotSupported"`
	}
	c.calculateSatisfied()
	item := &CheckPermissionsTypeAlias{
		Satisfied:            c.Satisfied,
		Permissions:          c.Permissions,
		ValidCombinations:    c.ValidCombinations,
		NetworksNotSupported: c.NetworksNotSupported,
	}
	rolesAndHighestRole := calculateRolesAndHighestRole(c.Permissions)

	item.Roles = rolesAndHighestRole.Roles
	item.HighestRole = rolesAndHighestRole.HighestRole
	return json.Marshal(item)
}

type TokenRequirementResponse struct {
	Satisfied     bool                    `json:"satisfied"`
	TokenCriteria *protobuf.TokenCriteria `json:"criteria"`
}

type PermissionTokenCriteriaResult struct {
	Role              protobuf.CommunityTokenPermission_Type `json:"roles"`
	TokenRequirements []TokenRequirementResponse             `json:"tokenRequirement"`
	Criteria          []bool                                 `json:"criteria"`
}

type AccountChainIDsCombination struct {
	Address  gethcommon.Address `json:"address"`
	ChainIDs []uint64           `json:"chainIds"`
}

func (c *CheckPermissionsResponse) calculateSatisfied() {
	if len(c.Permissions) == 0 {
		c.Satisfied = true
		return
	}

	c.Satisfied = false
	for _, p := range c.Permissions {
		satisfied := true
		for _, criteria := range p.Criteria {
			if !criteria {
				satisfied = false
				break
			}
		}
		if satisfied {
			c.Satisfied = true
			return
		}
	}
}
