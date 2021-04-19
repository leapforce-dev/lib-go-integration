package integration

import credentials "github.com/leapforce-libraries/go_google/credentials"

type Config struct {
	Name                  string
	AppName               string
	ProjectID             string
	Bucket                string
	Dataset               string
	SentryDSN             string
	ServiceAccountJSONKey *credentials.CredentialsJSON
}
