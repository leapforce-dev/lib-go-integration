package integration

import (
	"net/http"
)

var retryUponHttpStatusCode []int

func HttpRetry(httpStatusCode int) bool {
	for _, s := range retryUponHttpStatusCode {
		if s == httpStatusCode {
			return true
		}
	}

	return false
}

func SetHttpRetry(statusCodes []int) {
	retryUponHttpStatusCode = statusCodes
}

func AddHttpRetry(statusCodes []int) {
	retryUponHttpStatusCode = append(retryUponHttpStatusCode, statusCodes...)
}

func ResetHttpRetry() {
	initHttpRetry()
}

func initHttpRetry() {
	retryUponHttpStatusCode = []int{http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusGatewayTimeout}
}
