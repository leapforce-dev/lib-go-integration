package integration

import (
	"os"
	"strings"
)

var debug bool

func Debug() bool {
	return debug
}

func SetDebug(_debug bool) {
	debug = _debug
}

func initDebug() {
	SetDebug(strings.ToLower(strings.Trim(os.Getenv("DEBUG"), " ")) == "true")
}
