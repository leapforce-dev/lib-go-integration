package integration

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"sync"
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
	sync.RWMutex
	config            *Config
	configName        string
	run               string
	logCredentials    *credentials.CredentialsJson
	logger            *gcs.Logger
	validEnvironments *[]string
	validModes        *[]string
	includeCompanyIds *[]int64
	excludeCompanyIds *[]int64
	apiServices       []*ApiServiceWithKey
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

	integration.setToday(false)

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

func (i Integration) setToday(lock bool) {
	if lock {
		i.Lock()
		defer i.Unlock()
	}

	_today := civil.DateOf(time.Now())
	today = &_today
	_tomorrow := _today.AddDays(1)
	tomorrow = &_tomorrow
}

func (i Integration) SetToday() {
	i.setToday(true)
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
	return i.log("start_softwareclientlicense", nil, &companyId, &softwareClientLicenseGuid, nil, false)
}

func (i Integration) EndSoftwareClientLicense(companyId int64, softwareClientLicenseGuid string) *errortools.Error {
	return i.log("end_softwareclientlicense", nil, &companyId, &softwareClientLicenseGuid, nil, false)
}

func (i Integration) start() *errortools.Error {
	return i.log("start", nil, nil, nil, nil, false)
}

func (i Integration) end() *errortools.Error {
	return i.log("end", nil, nil, nil, nil, false)
}

func (i Integration) Log(operation string, companyId *int64, softwareClientLicenseGuid *string, data interface{}) *errortools.Error {
	return i.log(operation, nil, companyId, softwareClientLicenseGuid, data, true)
}

func (i Integration) log(operation string, key *string, companyId *int64, softwareClientLicenseGuid *string, data interface{}, lock bool) *errortools.Error {
	if lock {
		i.Lock()
		defer i.Unlock()
	}

	if companyId == nil {
		companyId = i.config.LogCompanyId
	}
	if i.logger == nil {
		return errortools.ErrorMessage("Logger not initialized")
	}

	var apis []ApiInfo

	for _, apiService := range i.apiServices {
		if apiService.ApiService == nil {
			continue
		}

		if key != nil {
			if *key != apiService.Key {
				continue
			}
		}

		apiKey := (*apiService.ApiService).ApiKey()
		if len(apiKey) > apiKeyTruncationLength {
			apiKey = apiKey[:apiKeyTruncationLength] + strings.Repeat("*", apiKeyTruncationLength)
		}

		apis = append(apis, ApiInfo{
			Name:      (*apiService.ApiService).ApiName(),
			Key:       apiKey,
			CallCount: (*apiService.ApiService).ApiCallCount(),
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
		SoftwareClientLicenseGuid: go_bigquery.StringToNullString(softwareClientLicenseGuid),
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
	i.Lock()
	defer i.Unlock()

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
	i.Lock()
	defer i.Unlock()

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
	i.Lock()
	defer i.Unlock()

	var a []*ApiServiceWithKey

	for j := range apiServices {
		a = append(a, &ApiServiceWithKey{ApiService: apiServices[j]})
	}

	i.apiServices = a
}

func (i *Integration) AddApiService(apiService *ApiService, sender string, user string) string {
	i.Lock()
	defer i.Unlock()

	key := uuid.NewString()

	i.apiServices = append(i.apiServices, &ApiServiceWithKey{
		Key:        key,
		Sender:     sender,
		User:       user,
		ApiService: apiService,
	})

	return key
}

func (i *Integration) RemoveApiService(key string) *errortools.Error {
	i.Lock()
	defer i.Unlock()

	var a []*ApiServiceWithKey

	for j := range i.apiServices {
		if i.apiServices[j].Key == key {
			e := i.log(i.apiServices[j].Sender, nil, nil, nil, i.apiServices[j].User, false)
			if e != nil {
				return e
			}

			continue
		}

		a = append(a, i.apiServices[j])
	}

	i.apiServices = a

	return nil
}

func (i *Integration) ResetApiServices() {
	i.Lock()
	defer i.Unlock()

	for _, apiService := range i.apiServices {
		if apiService.ApiService == nil {
			continue
		}

		(*apiService.ApiService).ApiReset()
	}
}
