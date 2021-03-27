package integration

import "strings"

var currentEnvironment string

type Environment string

const (
	EnvironmentTest Environment = "test"
	EnvironmentLive Environment = "live"
)

func IsEnvironmentTest() bool {
	return currentEnvironment == string(EnvironmentTest)
}

func IsEnvironmentLive() bool {
	return currentEnvironment == string(EnvironmentLive)
}

func IsEnvironment(environment string) bool {
	return strings.ToUpper(currentEnvironment) == strings.ToUpper(environment)
}

func CurrentEnvironment() string {
	return strings.ToUpper(currentEnvironment)
}
