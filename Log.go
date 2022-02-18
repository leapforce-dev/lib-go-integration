package integration

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"
)

type LogOLD struct {
	AppName        string
	Environment    string
	Mode           string
	Run            string
	Timestamp      time.Time
	Operation      string
	OrganisationID bigquery.NullInt64
	Data           json.RawMessage
}

type Log struct {
	AppName        string
	Environment    string
	Mode           string
	Run            string
	Timestamp      time.Time
	Operation      string
	OrganisationID bigquery.NullInt64
	Apis           []ApiInfo
	Data           string
}

type ApiInfo struct {
	Name      string
	Key       string
	CallCount int64
}
