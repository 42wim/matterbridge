package kbchat

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ListTeamMembers struct {
	Result ListTeamMembersResult `json:"result"`
	Error  Error             `json:"error"`
}

type ListTeamMembersResult struct {
	Members ListTeamMembersResultMembers `json:"members"`
}

type ListTeamMembersResultMembers struct {
	Owners  []ListMembersOutputMembersCategory `json:"owners"`
	Admins  []ListMembersOutputMembersCategory `json:"admins"`
	Writers []ListMembersOutputMembersCategory `json:"writers"`
	Readers []ListMembersOutputMembersCategory `json:"readers"`
}

type ListMembersOutputMembersCategory struct {
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

type ListUserMemberships struct {
	Result ListUserMembershipsResult `json:"result"`
	Error  Error             `json:"error"`
}

type ListUserMembershipsResult struct {
	Teams []ListUserMembershipsResultTeam `json:"teams"`
}

type ListUserMembershipsResultTeam struct {
	TeamName string `json:"fq_name"`
	IsImplicitTeam bool `json:"is_implicit_team"`
	IsOpenTeam bool `json:"is_open_team"`
	Role int `json:"role"`
	MemberCount int `json:"member_count"`
}

func (a *API) ListMembersOfTeam(teamName string) (ListTeamMembersResultMembers, error) {
	empty := ListTeamMembersResultMembers{}

	apiInput := fmt.Sprintf(`{"method": "list-team-memberships", "params": {"options": {"team": "%s"}}}`, teamName)
	cmd := a.runOpts.Command("team", "api")
	cmd.Stdin = strings.NewReader(apiInput)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return empty, fmt.Errorf("failed to call keybase team api: %v", err)
	}

	members := ListTeamMembers{}
	err = json.Unmarshal(bytes, &members)
	if err != nil {
		return empty, fmt.Errorf("failed to parse output from keybase team api: %v", err)
	}
	if members.Error.Message != "" {
		return empty, fmt.Errorf("received error from keybase team api: %s", members.Error.Message)
	}
	return members.Result.Members, nil
}

func (a *API) ListUserMemberships(username string) ([]ListUserMembershipsResultTeam, error) {
	empty := []ListUserMembershipsResultTeam{}

	apiInput := fmt.Sprintf(`{"method": "list-user-memberships", "params": {"options": {"username": "%s"}}}`, username)
	cmd := a.runOpts.Command("team", "api")
	cmd.Stdin = strings.NewReader(apiInput)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return empty, fmt.Errorf("failed to call keybase team api: %v", err)
	}

	members := ListUserMemberships{}
	err = json.Unmarshal(bytes, &members)
	if err != nil {
		return empty, fmt.Errorf("failed to parse output from keybase team api: %v", err)
	}
	if members.Error.Message != "" {
		return empty, fmt.Errorf("received error from keybase team api: %s", members.Error.Message)
	}
	return members.Result.Teams, nil
}
