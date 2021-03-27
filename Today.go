package integration

import (
	"cloud.google.com/go/civil"
)

var today *civil.Date

func TodayPtr() *civil.Date {
	return today
}

func Today() civil.Date {
	return *TodayPtr()
}
