// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

type GlobalRetentionPolicy struct {
	MessageDeletionEnabled bool  `json:"message_deletion_enabled"`
	FileDeletionEnabled    bool  `json:"file_deletion_enabled"`
	BoardsDeletionEnabled  bool  `json:"boards_deletion_enabled"`
	MessageRetentionCutoff int64 `json:"message_retention_cutoff"`
	FileRetentionCutoff    int64 `json:"file_retention_cutoff"`
	BoardsRetentionCutoff  int64 `json:"boards_retention_cutoff"`
}

type RetentionPolicy struct {
	ID           string `db:"Id" json:"id"`
	DisplayName  string `json:"display_name"`
	PostDuration *int64 `json:"post_duration"`
}

type RetentionPolicyWithTeamAndChannelIDs struct {
	RetentionPolicy
	TeamIDs    []string `json:"team_ids"`
	ChannelIDs []string `json:"channel_ids"`
}

type RetentionPolicyWithTeamAndChannelCounts struct {
	RetentionPolicy
	ChannelCount int64 `json:"channel_count"`
	TeamCount    int64 `json:"team_count"`
}

type RetentionPolicyChannel struct {
	PolicyID  string `db:"PolicyId"`
	ChannelID string `db:"ChannelId"`
}

type RetentionPolicyTeam struct {
	PolicyID string `db:"PolicyId"`
	TeamID   string `db:"TeamId"`
}

type RetentionPolicyWithTeamAndChannelCountsList struct {
	Policies   []*RetentionPolicyWithTeamAndChannelCounts `json:"policies"`
	TotalCount int64                                      `json:"total_count"`
}

type RetentionPolicyForTeam struct {
	TeamID       string `db:"Id" json:"team_id"`
	PostDuration int64  `json:"post_duration"`
}

type RetentionPolicyForTeamList struct {
	Policies   []*RetentionPolicyForTeam `json:"policies"`
	TotalCount int64                     `json:"total_count"`
}

type RetentionPolicyForChannel struct {
	ChannelID    string `db:"Id" json:"channel_id"`
	PostDuration int64  `json:"post_duration"`
}

type RetentionPolicyForChannelList struct {
	Policies   []*RetentionPolicyForChannel `json:"policies"`
	TotalCount int64                        `json:"total_count"`
}

type RetentionPolicyCursor struct {
	ChannelPoliciesDone bool
	TeamPoliciesDone    bool
	GlobalPoliciesDone  bool
}
