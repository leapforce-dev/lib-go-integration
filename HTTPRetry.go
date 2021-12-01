package integration

import (
	"net/http"
)

var retryUponHTTPStatusCode []int

func HTTPRetry(httpStatusCode int) bool {
	for _, s := range retryUponHTTPStatusCode {
		if s == httpStatusCode {
			return true
		}
	}

	return false
}

func SetHTTPRetry(statusCodes []int) {
	retryUponHTTPStatusCode = statusCodes
}

func AddHTTPRetry(statusCodes []int) {
	retryUponHTTPStatusCode = append(retryUponHTTPStatusCode, statusCodes...)
}

func ResetHTTPRetry() {
	initHTTPRetry()
}

func initHTTPRetry() {
	retryUponHTTPStatusCode = []int{http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusGatewayTimeout}
}
