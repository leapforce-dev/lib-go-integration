package integration

import (
	"fmt"

	errortools "github.com/leapforce-libraries/go_errortools"
	utilities "github.com/leapforce-libraries/go_utilities"
)

var currentEnvironment Environment
var currentMode Mode

type Integration struct {
	//environment Environment
	//mode        Mode
	validModes *[]Mode
}

type IntegrationConfig struct {
	ValidModes *[]Mode
}

func NewIntegration(integrationConfig *IntegrationConfig) (*Integration, *errortools.Error) {
	var environment string
	var mode string
	var arguments []*string

	arguments = append(arguments, &environment)
	if integrationConfig.ValidModes != nil {
		arguments = append(arguments, &mode)
	}

	e := utilities.GetArguments(arguments...)
	if e != nil {
		return nil, e
	}

	integration := Integration{
		//environment: environmentNone,
		//mode:        modeNone,
		validModes: integrationConfig.ValidModes,
	}

	if !integration.envIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid environment: '%s'", environment))
	}

	currentEnvironment = Environment(environment)

	if !integration.modeIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid mode: '%s'", mode))
	}

	currentMode = Mode(mode)

	return &integration, nil
}

func IsEnvTest() bool {
	return currentEnvironment == EnvironmentTest
}

func IsEnvLive() bool {
	return currentEnvironment == EnvironmentLive
}

func (Integration) envIsValid() bool {
	return IsEnvTest() || IsEnvLive()
}

func IsModeRecent() bool {
	return currentMode == ModeRecent
}

func IsModeHistory() bool {
	return currentMode == ModeHistory
}

func (i Integration) modeIsValid() bool {
	return IsModeRecent() || IsModeHistory()
}
