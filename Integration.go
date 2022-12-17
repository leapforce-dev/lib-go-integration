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
	logBucketName          string = "leapforce_xxx_log"
	logProjectId           string = "leapforce-224115"
	logDataset             string = "leapforce"
	apiKeyTruncationLength int    = 8
)

type Integration struct {
	config            *Config
	configName        string
	run               string
	logCredentials    *credentials.CredentialsJson
	logger            *gcs.Logger
	validEnvironments *[]string
	validModes        *[]string
	includeCompanyIds *[]int64
	excludeCompanyIds *[]int64
	apiServices       []*ApiService
}

type IntegrationConfig struct {
	DefaultConfig             *Config
	OtherConfigs              map[string]*Config
	LogCredentials            *credentials.CredentialsJson
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
	Arguments                 *[]string
}

func NewIntegration(integrationConfig *IntegrationConfig) (*Integration, *errortools.Error) {
	if integrationConfig == nil {
		return nil, errortools.ErrorMessage("IntegrationConfig is nil pointer")
	}

	initDebug()
	initHttpRetry()

	var validEnvironments, validModes = &[]string{}, &[]string{}

	var hasEnvironment, hasEnvironmentTest, hasEnvironmentLive = true, true, true
	var hasMode, hasModeRecent, hasModeHistory = true, true, true

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

	prefixArguments, e := utilities.GetArguments(&required, integrationConfig.Arguments, arguments...)
	if e != nil {
		return nil, e
	}

	// extract config
	var config *Config = nil
	var configName = "default"
	if prefixArguments != nil {
		_configName, ok := (*prefixArguments)["c"]
		if ok {
			_config, okConfig := integrationConfig.OtherConfigs[strings.ToLower(_configName)]
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

	var excludeCompanyIds *[]int64 = nil
	if len(integrationConfig.OtherConfigs) > 0 {
		excludeCompanyIds = new([]int64)
		for _, c := range integrationConfig.OtherConfigs {
			if c.CompanyIds != nil {
				*excludeCompanyIds = append((*excludeCompanyIds), *c.CompanyIds...)
			}
		}
	}

	guid := go_types.NewGuid()
	integration := Integration{
		config:            config,
		configName:        configName,
		run:               guid.String(),
		logCredentials:    integrationConfig.LogCredentials,
		validEnvironments: validEnvironments,
		validModes:        validModes,
		logger:            nil,
		includeCompanyIds: config.CompanyIds,
		excludeCompanyIds: excludeCompanyIds,
	}

	if !integration.environmentIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid environment: '%s'", CurrentEnvironment()))
	}

	if !integration.modeIsValid() {
		return nil, errortools.ErrorMessage(fmt.Sprintf("Invalid mode: '%s'", CurrentMode()))
	}

	integration.SetToday()

	//init logger
	e = integration.initLogger()
	if e != nil {
		return nil, e
	}

	//write start
	e = integration.start()
	if e != nil {
		return nil, e
	}

	return &integration, nil
}

func (i *Integration) initLogger() *errortools.Error {
	bucketName := logBucketName
	gcsServiceConfig := gcs.ServiceConfig{
		DefaultBucketName: &bucketName,
		CredentialsJson:   i.logCredentials,
	}
	gcsService, e := gcs.NewService(&gcsServiceConfig)
	if e != nil {
		return e
	}

	objectName := fmt.Sprintf("%s_%s", i.Config().AppName, time.Now().Format("20060102150405"))
	logger, e := gcsService.NewLogger(objectName, &Log{})
	if e != nil {
		return e
	}

	i.logger = logger

	return nil
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

func (i Integration) Config() *Config {
	return i.config
}

func (i Integration) ConfigName() string {
	return i.configName
}

func (i Integration) DoCompany(companyId int64) bool {
	if i.includeCompanyIds != nil {
		for _, o := range *i.includeCompanyIds {
			if o == companyId {
				return true
			}
		}

		return false
	}
	if i.excludeCompanyIds != nil {
		for _, o := range *i.excludeCompanyIds {
			if o == companyId {
				return false
			}
		}

		return true
	}

	return true
}

func (i Integration) SetToday() {
	_today := civil.DateOf(time.Now())
	today = &_today
	_tomorrow := _today.AddDays(1)
	tomorrow = &_tomorrow
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

func (i Integration) StartSoftwareClientLicense(companyId int64, softwareClientLicenseGuid string) *errortools.Error {
	return i.Log("start_softwareclientlicense", &companyId, &softwareClientLicenseGuid, nil)
}

func (i Integration) EndSoftwareClientLicense(companyId int64, softwareClientLicenseGuid string) *errortools.Error {
	return i.Log("end_softwareclientlicense", &companyId, &softwareClientLicenseGuid, nil)
}

func (i Integration) start(apiServices ...*ApiService) *errortools.Error {
	return i.Log("start", nil, nil, nil)
}

func (i Integration) end(apiServices ...*ApiService) *errortools.Error {
	return i.Log("end", nil, nil, nil)
}

func (i Integration) Log(operation string, companyId *int64, SoftwareClientLicenseGuid *string, data interface{}) *errortools.Error {
	if companyId == nil {
		companyId = i.config.LogCompanyId
	}
	if i.logger == nil {
		return errortools.ErrorMessage("Logger not initialized")
	}

	apis := []ApiInfo{}

	for _, apiService := range i.apiServices {
		if apiService == nil {
			continue
		}

		apiKey := (*apiService).ApiKey()
		if len(apiKey) > apiKeyTruncationLength {
			apiKey = apiKey[:apiKeyTruncationLength] + strings.Repeat("*", apiKeyTruncationLength)
		}

		apis = append(apis, ApiInfo{
			Name:      (*apiService).ApiName(),
			Key:       apiKey,
			CallCount: (*apiService).ApiCallCount(),
		})
	}

	log := Log{
		AppName:                   i.config.AppName,
		Environment:               CurrentEnvironment(),
		Mode:                      CurrentMode(),
		Run:                       i.run,
		Timestamp:                 time.Now(),
		Operation:                 operation,
		CompanyId:                 go_bigquery.Int64ToNullInt64(companyId),
		SoftwareClientLicenseGuid: go_bigquery.StringToNullString(SoftwareClientLicenseGuid),
		Apis:                      apis,
	}

	if !utilities.IsNil(data) {
		b, err := json.Marshal(data)
		if err != nil {
			return errortools.ErrorMessage(err)
		}
		log.Data = string(b)
	}

	return i.logger.Write(&log)
}

func (i *Integration) SaveLog(reInit bool) *errortools.Error {
	if i.logger == nil {
		return errortools.ErrorMessage("Logger not initialized")
	}

	e := i.logger.Close()
	if e != nil {
		return e
	}

	i.logger = nil

	if reInit {
		e = i.initLogger()
		if e != nil {
			return e
		}
	}

	return nil
}

func (i *Integration) Close() *errortools.Error {
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

func (i *Integration) ApiServices(apiServices ...*ApiService) {
	i.apiServices = apiServices
}

func (i *Integration) AddApiService(apiService *ApiService) {
	if apiService == nil {
		return
	}
	i.apiServices = append(i.apiServices, apiService)
}

func (i *Integration) RemoveApiService(apiService *ApiService) {
	if apiService == nil {
		return
	}

	var apiServices []*ApiService
	for j := range i.apiServices {
		if i.apiServices[j] == nil {
			continue
		}
		if (*i.apiServices[j]).ApiName() == (*apiService).ApiName() &&
			(*i.apiServices[j]).ApiKey() == (*apiService).ApiKey() {
			continue
		}
		apiServices = append(apiServices, i.apiServices[j])
	}

	i.apiServices = apiServices
}

func (i *Integration) ResetApiServices() {
	for _, apiService := range i.apiServices {
		if apiService == nil {
			continue
		}

		(*apiService).ApiReset()
	}
}
