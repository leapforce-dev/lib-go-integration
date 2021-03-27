package integration

import "strings"

type Table struct {
	Name    string
	Schema  interface{}
	Replace *TableReplace
	Merge   *TableMerge
}

type TableReplace struct {
	DateRangeField *string
	DateField      *string
}

type TableMerge struct {
	IDField string
}

// TableName returns tablename
//
func (t Table) TableName() string {
	return t.Name
}

// ObjectName returns ObjectName
//
func (t Table) ObjectName() string {
	objectName := t.Name

	if !IsEnvironmentLive() {
		objectName += strings.ToUpper(currentEnvironment)
	}

	return objectName
}
