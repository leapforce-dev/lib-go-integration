package integration

import (
	"time"

	"cloud.google.com/go/bigquery"
)

type Log struct {
	AppName                   string
	Environment               string
	Mode                      string
	Run                       string
	Timestamp                 time.Time
	Operation                 string
	CompanyId                 bigquery.NullInt64
	SoftwareClientLicenseGuid bigquery.NullString
	Apis                      []ApiInfo
	Data                      string
}

type ApiInfo struct {
	Name      string
	Key       string
	CallCount int64
}
