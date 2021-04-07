package integration

import (
	"fmt"
	"strings"
)

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

func IsMode(mode interface{}) bool {
	return strings.ToUpper(currentMode) == strings.ToUpper(fmt.Sprintf("%v", mode))
}

func modeIsValid() bool {
	return IsModeRecent() || IsModeHistory()
}

func CurrentMode() string {
	return strings.ToUpper(currentMode)
}
