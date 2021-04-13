package integration

import (
	"encoding/json"
	"time"
)

type Log struct {
	AppName     string
	Environment string
	Mode        string
	Run         string
	Timestamp   time.Time
	Operation   string
	Data        json.RawMessage
}
