package alias

import (
	"strings"
)

func IsAdjective(val string) bool {
	for _, v := range adjectives {
		if v == val {
			return true
		}
	}
	return false
}

func IsAnimal(val string) bool {
	for _, v := range animals {
		if v == val {
			return true
		}
	}
	return false
}

func IsAlias(alias string) bool {
	aliasParts := strings.Fields(alias)
	if len(aliasParts) == 3 {
		if IsAdjective(strings.Title(aliasParts[0])) && IsAdjective(strings.Title(aliasParts[1])) && IsAnimal(strings.Title(aliasParts[2])) {
			return true
		}
	}
	return false
}
