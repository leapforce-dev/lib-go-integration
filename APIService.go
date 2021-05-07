package integration

type APIService interface {
	APIName() string
	APICallCount() int64
}
