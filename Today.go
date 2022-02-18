package integration

import (
	"cloud.google.com/go/civil"
)

var today *civil.Date
var tomorrow *civil.Date

func TodayPtr() *civil.Date {
	return today
}

func Today() civil.Date {
	return *TodayPtr()
}

func TomorrowPtr() *civil.Date {
	return tomorrow
}

func Tomorrow() civil.Date {
	return *TomorrowPtr()
}
