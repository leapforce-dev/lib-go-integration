package integration

import (
	"fmt"

	credentials "github.com/leapforce-libraries/go_google/credentials"
)

type Config struct {
	AppName               string
	ProjectID             string
	Bucket                string
	Dataset               string
	SentryDSN             string
	settings              map[string]string
	ServiceAccountJSONKey *credentials.CredentialsJSON
	OrganisationIDs       *[]int64 // if nil, app runs for all organisationIDs not specified in any config from OtherConfigs
}

func (config *Config) Set(key interface{}, value string) {
	if config.settings == nil {
		config.settings = make(map[string]string)
	}

	config.settings[fmt.Sprintf("%v", key)] = value
}

func (config *Config) Get(key interface{}) (string, bool) {
	if config.settings == nil {
		return "", false
	}

	value, ok := config.settings[fmt.Sprintf("%v", key)]

	return value, ok
}
