package integration

import "strings"

type Table struct {
	objectName string
	Schema     interface{}
	Replace    *TableReplace
	Merge      *TableMerge
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
	return t.objectName
}

// ObjectName returns ObjectName
//
func (t Table) ObjectName() string {
	objectName := t.objectName

	if !IsEnvironmentLive() {
		objectName += strings.ToLower(currentEnvironment)
	}

	return objectName
}
