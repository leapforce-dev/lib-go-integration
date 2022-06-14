package integration

import (
	"fmt"

	credentials "github.com/leapforce-libraries/go_google/credentials"
)

type Config struct {
	AppName      string
	ProjectId    string
	Bucket       string
	Dataset      string
	SentryDsn    string
	settings     map[string]string
	Credentials  *credentials.CredentialsJson
	LogCompanyId *int64   // if the integration runs for a single company pass it's Id here
	CompanyIds   *[]int64 // if nil, app runs for all companyIds not specified in any config from OtherConfigs
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

func (config *Config) FullTableName(tableName string, quoted bool) string {
	_tableName := fmt.Sprintf("%s.%s.%s", config.ProjectId, config.Dataset, tableName)

	if quoted {
		_tableName = fmt.Sprintf("`%s`", _tableName)
	}

	return _tableName
}
