package integration

import (
	"net/http"
)

var retryUponHTTPStatusCode []int

func Retry(httpStatusCode int) bool {
	for _, s := range retryUponHTTPStatusCode {
		if s == httpStatusCode {
			return true
		}
	}

	return false
}

func SetRetry(statusCodes []int) {
	retryUponHTTPStatusCode = statusCodes
}

func AddRetry(statusCodes []int) {
	retryUponHTTPStatusCode = append(retryUponHTTPStatusCode, statusCodes...)
}

func ResetRetry() {
	initRetry()
}

func initRetry() {
	retryUponHTTPStatusCode = []int{http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusGatewayTimeout}
}
