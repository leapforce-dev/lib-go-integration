package integration

import credentials "github.com/leapforce-libraries/go_google/credentials"

type Config struct {
	AppName               string
	ProjectID             string
	Bucket                string
	Dataset               string
	SentryDSN             string
	ServiceAccountJSONKey *credentials.CredentialsJSON
}
