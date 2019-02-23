package bslack

import (
	"io/ioutil"
	"testing"

	"github.com/42wim/matterbridge/bridge"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestExtractTopicOrPurpose(t *testing.T) {
	testcases := map[string]struct {
		input          string
		wantChangeType string
		wantOutput     string
	}{
		"success - topic type":   {"@someone set channel topic: foo bar", "topic", "foo bar"},
		"success - purpose type": {"@someone set channel purpose: foo bar", "purpose", "foo bar"},
		"success - one line":     {"@someone set channel topic: foo bar", "topic", "foo bar"},
		"success - multi-line":   {"@someone set channel topic: foo\nbar", "topic", "foo\nbar"},
		"success - cleared":      {"@someone cleared channel topic", "topic", ""},
		"error - unhandled":      {"some unmatched message", "unknown", ""},
	}

	logger := logrus.New()
	logger.SetOutput(ioutil.Discard)
	cfg := &bridge.Config{Bridge: &bridge.Bridge{Log: logrus.NewEntry(logger)}}
	b := newBridge(cfg)
	for name, tc := range testcases {
		gotChangeType, gotOutput := b.extractTopicOrPurpose(tc.input)

		assert.Equalf(t, tc.wantChangeType, gotChangeType, "This testcase failed: %s", name)
		assert.Equalf(t, tc.wantOutput, gotOutput, "This testcase failed: %s", name)
	}
}
