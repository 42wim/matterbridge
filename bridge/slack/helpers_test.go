package bslack

import (
	"testing"
)

func TestExtractTopicOrPurpose(t *testing.T) {
	tables := []struct{
		input string
		changeType string
		output string
	}{
		{"@someone set channel topic: one liner",    "topic",   "one liner"},
		{"@someone set channel purpose: one liner",  "purpose", "one liner"},
		{"@someone set channel topic: multi\nliner", "topic",   "multi\nliner"},
		{"@someone cleared channel topic",           "topic",   ""},
		{"some unmatched message",                   "unknown", ""},
	}

	for _, table := range tables {
		changeType, output := extractTopicOrPurpose(table.input)
		if changeType != table.changeType || output != table.output {
			t.Error()
		}
	}
}
