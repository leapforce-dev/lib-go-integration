package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	errortools "github.com/leapforce-libraries/go_errortools"
	go_bigquery "github.com/leapforce-libraries/go_google/bigquery"
	credentials "github.com/leapforce-libraries/go_google/credentials"
	gcs "github.com/leapforce-libraries/go_googlecloudstorage"
	go_types "github.com/leapforce-libraries/go_types"
	utilities "github.com/leapforce-libraries/go_utilities"
)

const (
	logBucketName string = "leapforce_xxx_log"
	logProjectID  string = "leapforce-224115"
	logDataset    string = "leapforce"
	logTableName  string = "log"
)

type Integration struct {
	appName           string
	run               string
	logger            *gcs.Logger
	validEnvironments *[]string
	validModes        *[]string
	organisationID    *int64
}

type IntegrationConfig struct {
	AppName                   string
	CredentialsJSON           *credentials.CredentialsJSON
	OrganisationID            *int64 // if the integration runs for a single organisation pass it's ID here
	HasEnvironment            *bool  //default true
	HasEnvironmentTest        *bool  //default true
	HasEnvironmentLive        *bool  //default true
	OtherEnvironments         *[]string
	HasMode                   *bool //default true
	HasModeRecent             *bool //default true
	HasModeHistory            *bool //default true
	OtherModes                *[]string
	OtherCompulsoryArguments  *[]*string
	OtherFacultativeArguments *[]*string
}

func NewIntegration(integrationConfig *IntegrationConfig) (*Integration, *errortools.Error) {
	gcsServiceConfig := gcs.ServiceConfig{
		BucketName:      logBucketName,
		CredentialsJSON: integrationConfig.CredentialsJSON,
	}
	gcsService, e := gcs.NewService(&gcsServiceConfig)
	if e != nil {
		return nil, e
	}

	objectName := fmt.Sprintf("%s_%s", integrationConfig.AppName, time.Now().Format("20060102150405"))
	logger, e := gcsService.NewLogger(objectName, &Log{})
	if e != nil {
		return nil, e
	}

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

	guid := go_types.NewGUID()
	integration := Integration{
		appName:           integrationConfig.AppName,
		run:               guid.String(),
		logger:            logger,
		validEnvironments: validEnvironments,
		validModes:        validModes,
		organisationID:    integrationConfig.OrganisationID,
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

	//write start
	e = integration.start()
	if e != nil {
		return nil, e
	}

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

func (i Integration) StartOrganisation(organisationID int64) *errortools.Error {
	return i.Log("start_organisation", &organisationID, nil)
}

func (i Integration) EndOrganisation(organisationID int64) *errortools.Error {
	return i.Log("end_organisation", &organisationID, nil)
}

func (i Integration) start() *errortools.Error {
	return i.Log("start", nil, nil)
}

func (i Integration) end() *errortools.Error {
	return i.Log("end", nil, nil)
}

func (i Integration) Log(operation string, organisationID *int64, data interface{}) *errortools.Error {
	if organisationID == nil {
		organisationID = i.organisationID
	}

	log := Log{
		AppName:        i.appName,
		Environment:    CurrentEnvironment(),
		Mode:           CurrentMode(),
		Run:            i.run,
		Timestamp:      time.Now(),
		Operation:      operation,
		OrganisationID: go_bigquery.Int64ToNullInt64(organisationID),
	}

	if !utilities.IsNil(data) {
		b, err := json.Marshal(data)
		if err != nil {
			return errortools.ErrorMessage(err)
		}
		log.Data = b
	}

	return i.logger.Write(&log)
}

func (i Integration) Close() *errortools.Error {
	e := i.end()
	if e != nil {
		return e
	}

	e = i.logger.Close()
	if e != nil {
		return e
	}

	return nil
}
