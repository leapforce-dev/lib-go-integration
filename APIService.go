package integration

type ApiService interface {
	ApiName() string
	ApiKey() string
	ApiCallCount() int64
	ApiReset()
}

type ApiServiceWithKey struct {
	Key        string
	Sender     string
	User       string
	ApiService *ApiService
}
