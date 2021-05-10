package integration

type APIService interface {
	APIName() string
	APIKey() string
	APICallCount() int64
	APIReset()
}
