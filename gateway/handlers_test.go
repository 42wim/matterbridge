package gateway

import (
	"github.com/42wim/matterbridge/bridge"
	"github.com/42wim/matterbridge/bridge/config"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestIgnoreEvent(t *testing.T) {
	eventTests := map[string]struct {
		input  string
		dest   *bridge.Bridge
		output bool
	}{
		"avatar mattermost": {
			input:  config.EventAvatarDownload,
			dest:   &bridge.Bridge{Protocol: "mattermost"},
			output: false,
		},
		"avatar slack": {
			input:  config.EventAvatarDownload,
			dest:   &bridge.Bridge{Protocol: "slack"},
			output: true,
		},
		"avatar telegram": {
			input:  config.EventAvatarDownload,
			dest:   &bridge.Bridge{Protocol: "telegram"},
			output: false,
		},
	}
	gw := &Gateway{}
	for testname, testcase := range eventTests {
		output := gw.ignoreEvent(testcase.input, testcase.dest)
		assert.Equalf(t, testcase.output, output, "case '%s' failed", testname)
	}

}

func TestExtractNick(t *testing.T) {
	eventTests := map[string]struct {
		search         string
		extract        string
		username       string
		text           string
		resultUsername string
		resultText     string
	}{
		"test1": {
			search:         "fromgitter",
			extract:        "<(.*?)>\\s+",
			username:       "fromgitter",
			text:           "<userx> blahblah",
			resultUsername: "userx",
			resultText:     "blahblah",
		},
		"test2": {
			search: "<.*?bot>",
			//extract:        `\((.*?)\)\s+`,
			extract:        "\\((.*?)\\)\\s+",
			username:       "<matterbot>",
			text:           "(userx) blahblah (abc) test",
			resultUsername: "userx",
			resultText:     "blahblah (abc) test",
		},
	}
	//	gw := &Gateway{}
	for testname, testcase := range eventTests {
		resultUsername, resultText, _ := extractNick(testcase.search, testcase.extract, testcase.username, testcase.text)
		assert.Equalf(t, testcase.resultUsername, resultUsername, "case '%s' failed", testname)
		assert.Equalf(t, testcase.resultText, resultText, "case '%s' failed", testname)
	}

}
