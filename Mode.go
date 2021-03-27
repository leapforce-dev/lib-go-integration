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
	return strings.ToLower(currentMode) == strings.ToLower(mode)
}

func modeIsValid() bool {
	return IsModeRecent() || IsModeHistory()
}

func (m Mode) in(modes *[]Mode) bool {
	if modes == nil {
		return false
	}

	for _, mode := range *modes {
		if string(mode) == string(m) {
			return true
		}
	}

	return false
}
