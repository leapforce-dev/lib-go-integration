package integration

import (
	"encoding/json"
	"time"
)

type Log struct {
	AppName        string
	Environment    string
	Mode           string
	Run            string
	Timestamp      time.Time
	Operation      string
	OrganisationID int64
	Data           json.RawMessage
}
