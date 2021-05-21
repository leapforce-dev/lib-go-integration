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

func (tableReplace *TableReplace) AddWhere(fieldName string, operator string, valueExpression string) {
	tableReplace.wheres = append(tableReplace.wheres, where{fieldName, operator, valueExpression})
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
