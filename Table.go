package integration

import (
	"fmt"
	"strings"
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

type Where struct {
	FieldName       string
	Operator        string
	ValueExpression string
}

type TableReplace struct {
	DateRangeField *string
	DateField      *string
	Wheres         *[]Where
}

type TableMerge struct {
	JoinFields []string
}

type TableTruncate struct {
}

func (tableReplace *TableReplace) WhereString() *string {
	if tableReplace.Wheres == nil {
		return nil
	}

	whereStrings := []string{}

	for _, where := range *tableReplace.Wheres {
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
