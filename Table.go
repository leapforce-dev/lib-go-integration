package integration

type Table struct {
	Name    string
	Schema  interface{}
	Replace *TableReplace
	Merge   *TableMerge
}

type Where struct {
	FieldName       string
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
