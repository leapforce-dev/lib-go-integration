package integration

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
