package integration

import (
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	errortools "github.com/leapforce-libraries/go_errortools"
	utilities "github.com/leapforce-libraries/go_utilities"
)

type Integration struct {
	validEnvironments *[]string
	validModes        *[]string
}

type IntegrationConfig struct {
	HasEnvironment            *bool //default true
	HasEnvironmentTest        *bool //default true
	HasEnvironmentLive        *bool //default true
	OtherEnvironments         *[]string
	HasMode                   *bool //default true
	HasModeRecent             *bool //default true
	HasModeHistory            *bool //default true
	OtherModes                *[]string
	OtherCompulsoryArguments  *[]*string
	OtherFacultativeArguments *[]*string
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
	var required int = 0

	if integration.validModes != nil {
		arguments = append(arguments, &currentMode)
		required++
	}
	if integration.validEnvironments != nil {
		arguments = append(arguments, &currentEnvironment)
		required++
	}
	if integrationConfig != nil {
		if integrationConfig.OtherCompulsoryArguments != nil {
			for i := range *integrationConfig.OtherCompulsoryArguments {
				arguments = append(arguments, (*integrationConfig.OtherCompulsoryArguments)[i])
				required++
			}
		}
		if integrationConfig.OtherFacultativeArguments != nil {
			for i := range *integrationConfig.OtherFacultativeArguments {
				arguments = append(arguments, (*integrationConfig.OtherFacultativeArguments)[i])
			}
		}
	}

	if len(arguments) > 0 {
		e := utilities.GetArguments(&required, arguments...)
		if e != nil {
			return nil, e
		}
	}

	if !integration.environmentIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid environment: '%s'", CurrentEnvironment()))
	}

	if !integration.modeIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid mode: '%s'", CurrentMode()))
	}

	integration.SetToday()

	return &integration, nil
}

func (i Integration) Print() {
	if i.validModes != nil {
		fmt.Printf(">>> Mode : %s\n", CurrentMode())
	}
	if i.validEnvironments != nil {
		fmt.Printf(">>> Environment : %s\n", CurrentEnvironment())
	}
}

func (i Integration) SetToday() {
	date := civil.DateOf(time.Now())
	today = &date
}

func (i Integration) environmentIsValid() bool {
	if i.validEnvironments == nil {
		return true
	}

	for _, environment := range *i.validEnvironments {
		if strings.ToUpper(environment) == CurrentEnvironment() {
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
		if strings.ToUpper(mode) == CurrentMode() {
			return true
		}
	}

	return false
}
