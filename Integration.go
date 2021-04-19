package integration

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	config            *Config
	configName        string
	run               string
	logger            *gcs.Logger
	validEnvironments *[]string
	validModes        *[]string
	logOrganisationID *int64
	organisationID    *int64
}

type IntegrationConfig struct {
	DefaultConfig             *Config
	OtherConfigs              map[string]*Config
	LogCredentials            *credentials.CredentialsJSON
	LogOrganisationID         *int64 // if the integration runs for a single organisation pass it's ID here
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
	if integrationConfig == nil {
		return nil, errortools.ErrorMessage("IntegrationConfig is nil pointer")
	}

	//if integrationConfig.Config.ServiceAccountJSONKey == nil {
	//	return nil, errortools.ErrorMessage("ServiceAccountJSONKey is nil pointer")
	//}
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

	var arguments []*string
	var required int = 0

	if validModes != nil {
		arguments = append(arguments, &currentMode)
		required++
	}
	if validEnvironments != nil {
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

	//if len(arguments) > 0 {
	prefixArguments, e := utilities.GetArguments(&required, arguments...)
	if e != nil {
		return nil, e
	}

	// extract config
	var config *Config = nil
	var configName string = "default"
	if prefixArguments != nil {
		_configName, ok := (*prefixArguments)["c"]
		if ok {
			_config, okConfig := integrationConfig.OtherConfigs[strings.ToLower(configName)]
			if okConfig {
				config = _config
				configName = _configName
			}

			if config == nil {
				return nil, errortools.ErrorMessage(fmt.Sprintf("Config '%s' not found", configName))
			}
		}
	}
	if config == nil {
		if integrationConfig.DefaultConfig == nil {
			return nil, errortools.ErrorMessage("DefaultConfig is nil")
		}
		config = integrationConfig.DefaultConfig
	}

	// extract organisationID
	var organisationID int64
	if prefixArguments != nil {
		organisationIDString, ok := (*prefixArguments)["o"]
		if ok {
			_organisationID, err := strconv.ParseInt(organisationIDString, 10, 64)
			if err != nil {
				return nil, errortools.ErrorMessage("OrganisationID must be an integer")
			}
			organisationID = _organisationID
		}
	}

	gcsServiceConfig := gcs.ServiceConfig{
		BucketName:      logBucketName,
		CredentialsJSON: integrationConfig.LogCredentials,
	}
	gcsService, e := gcs.NewService(&gcsServiceConfig)
	if e != nil {
		return nil, e
	}

	objectName := fmt.Sprintf("%s_%s", config.AppName, time.Now().Format("20060102150405"))
	logger, e := gcsService.NewLogger(objectName, &Log{})
	if e != nil {
		return nil, e
	}

	guid := go_types.NewGUID()
	integration := Integration{
		config:            config,
		configName:        configName,
		run:               guid.String(),
		logger:            logger,
		validEnvironments: validEnvironments,
		validModes:        validModes,
		logOrganisationID: integrationConfig.LogOrganisationID,
		organisationID:    &organisationID,
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
	fmt.Printf(">>> Config : %s\n", i.configName)
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
		organisationID = i.logOrganisationID
	}

	log := Log{
		AppName:        i.config.AppName,
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
