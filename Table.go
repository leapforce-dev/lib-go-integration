package integration

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
