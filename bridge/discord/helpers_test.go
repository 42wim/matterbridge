package bdiscord

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnumerateUsernames(t *testing.T) {
	testcases := map[string]struct {
		match             string
		expectedUsernames []string
	}{
		"only space": {
			match:             "  \t\n \t",
			expectedUsernames: nil,
		},
		"single word": {
			match:             "veni",
			expectedUsernames: []string{"veni"},
		},
		"single word with preceeding space": {
			match:             " vidi",
			expectedUsernames: []string{" vidi"},
		},
		"single word with suffixed space": {
			match:             "vici ",
			expectedUsernames: []string{"vici"},
		},
		"multi-word with varying whitespace": {
			match: "just me  and\tmy friends \t",
			expectedUsernames: []string{
				"just",
				"just me",
				"just me  and",
				"just me  and\tmy",
				"just me  and\tmy friends",
			},
		},
	}

	for testname, testcase := range testcases {
		foundUsernames := enumerateUsernames(testcase.match)
		assert.Equalf(t, testcase.expectedUsernames, foundUsernames, "Should have found the expected usernames for testcase %s", testname)
	}
}
