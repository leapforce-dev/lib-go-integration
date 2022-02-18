package integration

type ApiService interface {
	ApiName() string
	ApiKey() string
	ApiCallCount() int64
	ApiReset()
}
