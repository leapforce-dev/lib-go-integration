package integration

import (
	"cloud.google.com/go/civil"
)

var today *civil.Date
var tomorrow *civil.Date
var minDate *civil.Date = &civil.Date{Year: 1800, Month: 1, Day: 1}
var maxDate *civil.Date = &civil.Date{Year: 2099, Month: 12, Day: 31}

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

func MinDatePtr() *civil.Date {
	return minDate
}

func MinDate() civil.Date {
	return *minDate
}

func MaxDatePtr() *civil.Date {
	return maxDate
}

func MaxDate() civil.Date {
	return *maxDate
}
