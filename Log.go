package integration

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"
)

type Log struct {
	AppName        string
	Environment    string
	Mode           string
	Run            string
	Timestamp      time.Time
	Operation      string
	OrganisationID bigquery.NullInt64
	Data           json.RawMessage
}
