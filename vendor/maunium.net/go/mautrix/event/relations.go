// Copyright (c) 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package event

import (
	"encoding/json"

	"maunium.net/go/mautrix/id"
)

type RelationType string

const (
	RelReplace    RelationType = "m.replace"
	RelReference  RelationType = "m.reference"
	RelAnnotation RelationType = "m.annotation"
	RelReply      RelationType = "net.maunium.reply"
)

type RelatesTo struct {
	Type    RelationType
	EventID id.EventID
	Key     string
}

type serializableInReplyTo struct {
	EventID id.EventID `json:"event_id,omitempty"`
}

type serializableRelatesTo struct {
	InReplyTo *serializableInReplyTo `json:"m.in_reply_to,omitempty"`

	Type    RelationType `json:"rel_type,omitempty"`
	EventID id.EventID   `json:"event_id,omitempty"`
	Key     string       `json:"key,omitempty"`
}

func (rel *RelatesTo) GetReplaceID() id.EventID {
	if rel.Type == RelReplace {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetReferenceID() id.EventID {
	if rel.Type == RelReference {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetReplyID() id.EventID {
	if rel.Type == RelReply {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetAnnotationID() id.EventID {
	if rel.Type == RelAnnotation {
		return rel.EventID
	}
	return ""
}

func (rel *RelatesTo) GetAnnotationKey() string {
	if rel.Type == RelAnnotation {
		return rel.Key
	}
	return ""
}

func (rel *RelatesTo) UnmarshalJSON(data []byte) error {
	var srel serializableRelatesTo
	if err := json.Unmarshal(data, &srel); err != nil {
		return err
	}
	if len(srel.Type) > 0 {
		rel.Type = srel.Type
		rel.EventID = srel.EventID
		rel.Key = srel.Key
	} else if srel.InReplyTo != nil && len(srel.InReplyTo.EventID) > 0 {
		rel.Type = RelReply
		rel.EventID = srel.InReplyTo.EventID
		rel.Key = ""
	}
	return nil
}

func (rel *RelatesTo) MarshalJSON() ([]byte, error) {
	srel := serializableRelatesTo{Type: rel.Type, EventID: rel.EventID, Key: rel.Key}
	if rel.Type == RelReply {
		srel.InReplyTo = &serializableInReplyTo{rel.EventID}
	}
	return json.Marshal(&srel)
}

type RelationChunkItem struct {
	Type    RelationType `json:"type"`
	EventID string       `json:"event_id,omitempty"`
	Key     string       `json:"key,omitempty"`
	Count   int          `json:"count,omitempty"`
}

type RelationChunk struct {
	Chunk []RelationChunkItem `json:"chunk"`

	Limited bool `json:"limited"`
	Count   int  `json:"count"`
}

type AnnotationChunk struct {
	RelationChunk
	Map map[string]int `json:"-"`
}

type serializableAnnotationChunk AnnotationChunk

func (ac *AnnotationChunk) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*serializableAnnotationChunk)(ac)); err != nil {
		return err
	}
	ac.Map = make(map[string]int)
	for _, item := range ac.Chunk {
		ac.Map[item.Key] += item.Count
	}
	return nil
}

func (ac *AnnotationChunk) Serialize() RelationChunk {
	ac.Chunk = make([]RelationChunkItem, len(ac.Map))
	i := 0
	for key, count := range ac.Map {
		ac.Chunk[i] = RelationChunkItem{
			Type:  RelAnnotation,
			Key:   key,
			Count: count,
		}
	}
	return ac.RelationChunk
}

type EventIDChunk struct {
	RelationChunk
	List []string `json:"-"`
}

type serializableEventIDChunk EventIDChunk

func (ec *EventIDChunk) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*serializableEventIDChunk)(ec)); err != nil {
		return err
	}
	for _, item := range ec.Chunk {
		ec.List = append(ec.List, item.EventID)
	}
	return nil
}

func (ec *EventIDChunk) Serialize(typ RelationType) RelationChunk {
	ec.Chunk = make([]RelationChunkItem, len(ec.List))
	for i, eventID := range ec.List {
		ec.Chunk[i] = RelationChunkItem{
			Type:    typ,
			EventID: eventID,
		}
	}
	return ec.RelationChunk
}

type Relations struct {
	Raw map[RelationType]RelationChunk `json:"-"`

	Annotations AnnotationChunk `json:"m.annotation,omitempty"`
	References  EventIDChunk    `json:"m.reference,omitempty"`
	Replaces    EventIDChunk    `json:"m.replace,omitempty"`
}

type serializableRelations Relations

func (relations *Relations) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &relations.Raw); err != nil {
		return err
	}
	return json.Unmarshal(data, (*serializableRelations)(relations))
}

func (relations *Relations) MarshalJSON() ([]byte, error) {
	if relations.Raw == nil {
		relations.Raw = make(map[RelationType]RelationChunk)
	}
	relations.Raw[RelAnnotation] = relations.Annotations.Serialize()
	relations.Raw[RelReference] = relations.References.Serialize(RelReference)
	relations.Raw[RelReplace] = relations.Replaces.Serialize(RelReplace)
	return json.Marshal(relations.Raw)
}
