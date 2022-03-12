package bdiscord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestHandleEmbed(t *testing.T) {
	testcases := map[string]struct {
		embed  *discordgo.MessageEmbed
		result string
	}{
		"allempty": {
			embed:  &discordgo.MessageEmbed{},
			result: "",
		},
		"one": {
			embed: &discordgo.MessageEmbed{
				Title: "blah",
			},
			result: " embed: blah\n",
		},
		"two": {
			embed: &discordgo.MessageEmbed{
				Title:       "blah",
				Description: "blah2",
			},
			result: " embed: blah - blah2\n",
		},
		"three": {
			embed: &discordgo.MessageEmbed{
				Title:       "blah",
				Description: "blah2",
				URL:         "blah3",
			},
			result: " embed: blah - blah2 - blah3\n",
		},
		"twob": {
			embed: &discordgo.MessageEmbed{
				Description: "blah2",
				URL:         "blah3",
			},
			result: " embed: blah2 - blah3\n",
		},
		"oneb": {
			embed: &discordgo.MessageEmbed{
				URL: "blah3",
			},
			result: " embed: blah3\n",
		},
	}

	for name, tc := range testcases {
		assert.Equalf(t, tc.result, handleEmbed(tc.embed), "Testcases %s", name)
	}
}
