package integration

import (
	"fmt"
	"strings"

	"cloud.google.com/go/civil"
)

type Granularity string

const (
	GranularityNone  Granularity = ""
	GranularityDay   Granularity = "day"
	GranularityWeek  Granularity = "week"
	GranularityMonth Granularity = "month"
)

type Table struct {
	Name        string
	Granularity Granularity
	Schema      interface{}
	Replace     *TableReplace
	Merge       *TableMerge
	Truncate    *TableTruncate
}

type where struct {
	FieldName       string
	Operator        string
	ValueExpression string
}

type TableReplace struct {
	wheres []where
}

type TableMerge struct {
	JoinFields []string
}

type TableTruncate struct {
}

func (tableReplace *TableReplace) AddDummy() {
	tableReplace.wheres = append(tableReplace.wheres, where{"1", "=", "1"})
}

func (tableReplace *TableReplace) AddWhere(fieldName string, operator string, valueExpression string) {
	tableReplace.wheres = append(tableReplace.wheres, where{fieldName, operator, valueExpression})
}

func (tableReplace *TableReplace) AddWhereDate(fieldName string, date civil.Date) {
	tableReplace.wheres = append(tableReplace.wheres, where{fieldName, "=", fmt.Sprintf("'%s'", date.String())})
}

func (tableReplace *TableReplace) AddWhereDateRange(fieldName string, startDate civil.Date, endDate civil.Date) {
	tableReplace.wheres = append(tableReplace.wheres, where{fieldName, "BETWEEN", fmt.Sprintf("'%s' AND '%s'", startDate.String(), endDate.String())})
}

func (tableReplace *TableReplace) AddWhereDates(fieldName string, dates []civil.Date) {
	if len(dates) == 0 {
		return
	}

	var datesString []string

	for _, date := range dates {
		datesString = append(datesString, date.String())
	}

	tableReplace.wheres = append(tableReplace.wheres, where{fieldName, "IN", fmt.Sprintf("('%s')", strings.Join(datesString, "','"))})
}

func (tableReplace *TableReplace) WhereString() *string {
	if len(tableReplace.wheres) == 0 {
		return nil
	}

	whereStrings := []string{}

	for _, where := range tableReplace.wheres {
		fieldName := strings.Trim(where.FieldName, " ")
		if fieldName == "" {
			continue
		}
		operator := strings.Trim(where.Operator, " ")
		if operator == "" {
			operator = "="
		}
		valueExpression := strings.Trim(where.ValueExpression, " ")
		if valueExpression == "" {
			continue
		}

		whereStrings = append(whereStrings, fmt.Sprintf("%s %s %s", fieldName, operator, valueExpression))
	}

	if len(whereStrings) == 0 {
		return nil
	}

	whereString := strings.Join(whereStrings, " AND ")
	return &whereString
}
