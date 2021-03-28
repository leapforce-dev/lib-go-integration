package integration

import (
	"fmt"
	"strings"
)

var currentEnvironment string

type Environment string

const (
	EnvironmentTest Environment = "TEST"
	EnvironmentLive Environment = "LIVE"
)

func IsEnvironmentTest() bool {
	return CurrentEnvironment() == string(EnvironmentTest)
}

func IsEnvironmentLive() bool {
	return CurrentEnvironment() == string(EnvironmentLive)
}

func IsEnvironment(environment string) bool {
	return CurrentEnvironment() == strings.ToUpper(environment)
}

func CurrentEnvironment() string {
	return strings.ToUpper(currentEnvironment)
}

func WithEnvironment(names ...*string) {
	if IsEnvironmentLive() {
		return
	}

	for _, name := range names {
		if name == nil {
			continue
		}
		(*name) = fmt.Sprintf("%s_%s", *name, CurrentEnvironment())
	}
	return
}
