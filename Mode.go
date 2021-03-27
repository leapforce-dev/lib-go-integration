package integration

import "strings"

var currentMode string

type Mode string

const (
	ModeRecent  Mode = "recent"
	ModeHistory Mode = "history"
)

func IsModeRecent() bool {
	return currentMode == string(ModeRecent)
}

func IsModeHistory() bool {
	return currentMode == string(ModeHistory)
}

func IsMode(mode string) bool {
	return strings.ToUpper(currentMode) == strings.ToUpper(mode)
}

func modeIsValid() bool {
	return IsModeRecent() || IsModeHistory()
}

func CurrentMode() string {
	return strings.ToUpper(currentMode)
}
