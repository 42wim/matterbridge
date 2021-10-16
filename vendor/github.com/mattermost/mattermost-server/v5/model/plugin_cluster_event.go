// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

const (
	PluginClusterEventSendTypeReliable   = CLUSTER_SEND_RELIABLE
	PluginClusterEventSendTypeBestEffort = CLUSTER_SEND_BEST_EFFORT
)

// PluginClusterEvent is used to allow intra-cluster plugin communication.
type PluginClusterEvent struct {
	// Id is the unique identifier for the event.
	Id string
	// Data is the event payload.
	Data []byte
}

// PluginClusterEventSendOptions defines some properties that apply when sending
// plugin events across a cluster.
type PluginClusterEventSendOptions struct {
	// SendType defines the type of communication channel used to send the event.
	SendType string
	// TargetId identifies the cluster node to which the event should be sent.
	// It should match the cluster id of the receiving instance.
	// If empty, the event gets broadcasted to all other nodes.
	TargetId string
}
