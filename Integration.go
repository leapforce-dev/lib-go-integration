package integration

import (
	"fmt"
	"strings"

	errortools "github.com/leapforce-libraries/go_errortools"
	utilities "github.com/leapforce-libraries/go_utilities"
)

type Integration struct {
	validEnvironments *[]string
	validModes        *[]string
}

type IntegrationConfig struct {
	HasEnvironment     *bool //default true
	HasEnvironmentTest *bool //default true
	HasEnvironmentLive *bool //default true
	OtherEnvironments  *[]string
	HasMode            *bool //default true
	HasModeRecent      *bool //default true
	HasModeHistory     *bool //default true
	OtherModes         *[]string
}

func NewIntegration(integrationConfig *IntegrationConfig) (*Integration, *errortools.Error) {
	var validEnvironments, validModes *[]string = &[]string{}, &[]string{}

	var hasEnvironment, hasEnvironmentTest, hasEnvironmentLive bool = true, true, true
	var hasMode, hasModeRecent, hasModeHistory bool = true, true, true

	if integrationConfig != nil {
		if integrationConfig.HasEnvironment != nil {
			hasEnvironment = *integrationConfig.HasEnvironment
		}
		if hasEnvironment {
			if integrationConfig.HasEnvironmentTest != nil {
				hasEnvironmentTest = *integrationConfig.HasEnvironmentTest
			}
			if integrationConfig.HasEnvironmentLive != nil {
				hasEnvironmentLive = *integrationConfig.HasEnvironmentLive
			}
			if integrationConfig.OtherEnvironments != nil {
				validEnvironments = integrationConfig.OtherEnvironments
			}
		}

		if integrationConfig.HasMode != nil {
			hasMode = *integrationConfig.HasMode
		}
		if hasMode {
			if integrationConfig.HasModeRecent != nil {
				hasModeRecent = *integrationConfig.HasModeRecent
			}
			if integrationConfig.HasModeHistory != nil {
				hasModeHistory = *integrationConfig.HasModeHistory
			}
			if integrationConfig.OtherModes != nil {
				validModes = integrationConfig.OtherModes
			}
		}
	}

	if hasEnvironment {
		if hasEnvironmentTest {
			*validEnvironments = append(*validEnvironments, string(EnvironmentTest))
		}
		if hasEnvironmentLive {
			*validEnvironments = append(*validEnvironments, string(EnvironmentLive))
		}
	} else {
		validEnvironments = nil
	}

	if hasMode {
		if hasModeRecent {
			*validModes = append(*validModes, string(ModeRecent))
		}
		if hasModeHistory {
			*validModes = append(*validModes, string(ModeHistory))
		}
	} else {
		validModes = nil
	}

	integration := Integration{
		validEnvironments: validEnvironments,
		validModes:        validModes,
	}

	var arguments []*string

	if integration.validModes != nil {
		arguments = append(arguments, &currentMode)
	}
	if integration.validEnvironments != nil {
		arguments = append(arguments, &currentEnvironment)
	}

	if len(arguments) > 0 {
		e := utilities.GetArguments(arguments...)
		if e != nil {
			return nil, e
		}
	}

	if !integration.environmentIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid environment: '%s'", currentEnvironment))
	}

	if !integration.modeIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid mode: '%s'", currentMode))
	}

	return &integration, nil
}

func (i Integration) Print() {
	if i.validModes != nil {
		fmt.Printf(">>> Mode : %s\n", strings.ToLower(currentMode))
	}
	if i.validEnvironments != nil {
		fmt.Printf(">>> Environment : %s\n", strings.ToLower(currentEnvironment))
	}
}

func (i Integration) environmentIsValid() bool {
	if i.validEnvironments == nil {
		return true
	}

	for _, environment := range *i.validEnvironments {
		if strings.ToLower(environment) == strings.ToLower(currentEnvironment) {
			return true
		}
	}

	return false
}

func (i Integration) modeIsValid() bool {
	if i.validModes == nil {
		return true
	}

	for _, mode := range *i.validModes {
		if strings.ToLower(mode) == strings.ToLower(currentMode) {
			return true
		}
	}

	return false
}
