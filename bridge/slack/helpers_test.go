package bslack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTopicOrPurpose(t *testing.T) {
	testcases := []struct{
		input string
		wantChangeType string
		wantOutput string
	}{
		{"@someone set channel topic: one liner",    "topic",   "one liner"},
		{"@someone set channel purpose: one liner",  "purpose", "one liner"},
		{"@someone set channel topic: multi\nliner", "topic",   "multi\nliner"},
		{"@someone cleared channel topic",           "topic",   ""},
		{"some unmatched message",                   "unknown", ""},
	}

	for _, tc := range testcases {
		gotChangeType, gotOutput := extractTopicOrPurpose(tc.input)

		assert.Equal(t, tc.wantChangeType , gotChangeType)
		assert.Equal(t, tc.wantOutput , gotOutput)
	}
}
