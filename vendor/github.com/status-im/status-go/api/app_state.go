package api

import (
	"fmt"
	"strings"
)

// appState represents if the app is in foreground, background or some other state
type appState string

func (a appState) String() string {
	return string(a)
}

// Specific app states
// see https://facebook.github.io/react-native/docs/appstate.html
const (
	appStateForeground = appState("active") // these constant values are kept in sync with React Native
	appStateBackground = appState("background")
	appStateInactive   = appState("inactive")

	appStateInvalid = appState("")
)

// validAppStates returns an immutable set of valid states.
func validAppStates() []appState {
	return []appState{appStateInactive, appStateBackground, appStateForeground}
}

// parseAppState creates AppState from a string
func parseAppState(stateString string) (appState, error) {
	// a bit of cleaning up
	stateString = strings.ToLower(strings.TrimSpace(stateString))

	for _, state := range validAppStates() {
		if stateString == state.String() {
			return state, nil
		}
	}

	return appStateInvalid, fmt.Errorf("could not parse app state: %s", stateString)
}
