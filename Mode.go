package integration

import (
	"fmt"
	"strings"
)

var currentMode string

type Mode string

const (
	ModeRecent  Mode = "RECENT"
	ModeHistory Mode = "HISTORY"
)

func IsModeRecent() bool {
	return currentMode == string(ModeRecent)
}

func IsModeHistory() bool {
	return currentMode == string(ModeHistory)
}

func IsMode(mode interface{}) bool {
	return CurrentMode() == strings.ToUpper(fmt.Sprintf("%v", mode))
}

func CurrentMode() string {
	return strings.ToUpper(currentMode)
}
