package builder

import (
	"strconv"
	"strings"
)

var operators = []string{
	"=", "<", ">", "<=", ">=", "<>", "!=", "<=>",
	"like", "like binary", "not like", "ilike",
	"&", "|", "^", "<<", ">>",
	"rlike", "regexp", "not regexp",
	"~", "~*", "!~", "!~*", "similar to",
	"not similar to", "not ilike", "~~*", "!~~*"}

// Builder builder
type Builder struct {
	Connection       Connector     // The database connection instance.
	PSql             string        // Prepared sql
	PArgs            []interface{} // Prepared args
	Bindings         *Bindings     // The current query value bindings.
	Aggregate        map[string]string
	Columns          []string            // The columns that should be returned.
	Values           []map[string]string // update  and insert args
	IsDistinct       bool                // Indicates if the query returns distinct results.
	FromTable        string              // The table which the query is targeting.
	Groups           []string            // The grouping for the query.
	LimitNum         int                 // The maximum number of records to return.
	OffsetNum        int                 // The number of records to skip.
	Unions           map[string]interface{}
	UnionLimitNum    int
	UnionOffsetNum   int
	UnionOrders      map[string]interface{}
	Lock             map[string]interface{}
	Operators        map[string]interface{}
	Components       map[string][]map[string]string // compile Components
	SelectComponents []string                       // just for compile Component in order
	UseWrite         bool                           // Whether use write DB for select.
	debug            bool
}

// Bindings The current query value bindings.
type Bindings struct {
	Wheres []interface{}
	Having []interface{}
	Order  []interface{}
	Union  []interface{}
}

// New return a Builder
func New(c Connector) *Builder {
	return &Builder{
		Connection: c,
		Bindings:   &Bindings{},
		Components: map[string][]map[string]string{},
		SelectComponents: []string{
			"aggregate",
			"columns",
			"from",
			"joins",
			"wheres",
			"groups",
			"havings",
			"orders",
			"limit",
			"offset",
			"unions",
			"lock",
		}}
}

// Reset resets the Builder to be empty but
// keep Connection PSql PArgs and SelectComponents
func (b *Builder) Reset() {
	b.Bindings = &Bindings{}
	b.Aggregate = map[string]string{}
	b.Columns = []string{}
	b.Values = []map[string]string{}
	b.IsDistinct = false
	b.FromTable = ""
	b.Groups = []string{}
	b.LimitNum = 0
	b.OffsetNum = 0
	b.Unions = map[string]interface{}{}
	b.UnionLimitNum = 0
	b.UnionOffsetNum = 0
	b.UnionOrders = map[string]interface{}{}
	b.Lock = map[string]interface{}{}
	b.Operators = map[string]interface{}{}
	b.Components = map[string][]map[string]string{}
	b.UseWrite = false
}

// Debug debug mode
func (b *Builder) Debug(bo ...bool) *Builder {
	b.debug = true
	if bo != nil && bo[0] == false {
		b.debug = false
	}
	return b
}

// Insert a new record into the database. CURD [C]
func (b *Builder) Insert(values []map[string]string) int64 {
	b.Values = values
	return b.Connection.Insert()
}

// Update a record in the database. CURD [U]
func (b *Builder) Update(value map[string]string) int64 {

	b.prepareBindingsForUpdate(value)

	// 返回受影响的行数
	return b.Connection.Update()
}

// Increment a column's value by a given amount.
func (b *Builder) Increment(column string, amount ...string) bool {
	num := " + 1"
	if amount != nil {
		num = " + " + amount[0]
	}

	columns := map[string]string{
		column: column + num,
	}

	return b.Update(columns) > 0
}

// Prepare the bindings for an update statement.
func (b *Builder) prepareBindingsForUpdate(value map[string]string) {
	b.Values = append(b.Values, value)

	tem := make([]interface{}, len(b.Bindings.Wheres))
	copy(tem, b.Bindings.Wheres)
	b.Bindings.Wheres = nil

	for _, v := range value {
		b.Bindings.Wheres = append(b.Bindings.Wheres, v)
	}

	b.Bindings.Wheres = append(b.Bindings.Wheres, tem...)
}

// Exists Run the query as a "Exists" statement
func (b *Builder) Exists() bool {
	if b.Columns == nil {
		b.Columns = append(b.Columns, "*")
		b.Components["columns"] = nil
	}

	return b.Connection.Exists()
}

// RunSelect Run the query as a "select" statement against the connection.CURD [R]
func (b *Builder) RunSelect() []map[string]interface{} { return b.Connection.Select(true) }

// Delete a record from the database. CURD [D]
func (b *Builder) Delete() int64 { return b.Connection.Delete() }

// UseWriteDB Use the write DB for query.
func (b *Builder) UseWriteDB() { b.UseWrite = true }

// Determine if the given operator and value combination is legal.
func invalidOperatorAndValue(operator, value string) bool {
	hasop, nc := false, false
	for _, v := range operators {
		if operator == v {
			hasop = true
		}
	}
	for _, v := range []string{"=", "<>", "!="} {
		if operator != v {
			nc = true
		}
	}
	return value == "" && hasop && nc
}

func (b *Builder) invalidOperator(operator string) bool {
	op := strings.ToLower(operator)
	inOp := true
	for _, v := range operators {
		if op == v {
			inOp = false
		}
	}
	return inOp
}

// GroupBy group by
func (b *Builder) GroupBy(column string) *Builder {
	b.Groups = append(b.Groups, column)
	b.Components["groups"] = nil

	return b
}

// OrderBy order by
func (b *Builder) OrderBy(column string, direction ...string) *Builder {
	if direction == nil {
		direction = append(direction, "asc")
	} else {
		direction[0] = strings.ToLower(direction[0])
	}

	order := map[string]string{
		"column":    column,
		"direction": direction[0],
	}

	b.Components["orders"] = append(b.Components["orders"], order)
	return b
}

// OrderByDesc desc
func (b *Builder) OrderByDesc(column string) *Builder { return b.OrderBy(column, "desc") }

// WhereNotBetween where no between
func (b *Builder) WhereNotBetween(column string, values ...string) *Builder {

	return b.whereBetweenOrIn("Between", column, values, "and", true)
}

// OrWhereNotBetween or
func (b *Builder) OrWhereNotBetween(column string, values ...string) *Builder {

	return b.whereBetweenOrIn("Between", column, values, "or", true)
}

// WhereBetween where between
func (b *Builder) WhereBetween(column string, values ...string) *Builder {

	return b.whereBetweenOrIn("Between", column, values, "and")
}

// OrWhereBetween or
func (b *Builder) OrWhereBetween(column string, values ...string) *Builder {

	return b.whereBetweenOrIn("Between", column, values, "or")
}

// WhereIn in
func (b *Builder) WhereIn(column string, values ...string) *Builder {
	return b.whereBetweenOrIn("In", column, values, "and")
}

// OrWhereIn or in
func (b *Builder) OrWhereIn(column string, values ...string) *Builder {
	return b.whereBetweenOrIn("In", column, values, "or")
}

// WhereNotIn not in
func (b *Builder) WhereNotIn(column string, values ...string) *Builder {
	return b.whereBetweenOrIn("In", column, values, "and", true)
}

// OrWhereNotIn or not in
func (b *Builder) OrWhereNotIn(column string, values ...string) *Builder {
	return b.whereBetweenOrIn("In", column, values, "or", true)
}

func (b *Builder) whereBetweenOrIn(whereDateType, column string, values []string, logical string, isNot ...bool) *Builder {
	var not bool
	if isNot != nil && isNot[0] {
		not = true
	}

	where := map[string]string{
		"type":    whereDateType,
		"column":  column,
		"not":     strconv.FormatBool(not),
		"logical": logical,
		"value":   strings.Join(values, ","),
	}

	b.Components["wheres"] = append(b.Components["wheres"], where)

	// Binding values
	b.addBindings(where)

	return b
}

func dealValues(values ...string) string {
	var value string
	if len(values) > 0 {
		value = values[0]
	}
	return value
}

// dealWhereValues
// Where("age", ">=", "22", "(")
// example value = 22, pt= "("
func dealWhereValues(values ...string) (value, pt string) {
	if len(values) > 1 {
		pt = values[1]
	}

	if len(values) > 0 {
		value = values[0]
	}

	return
}

// WhereTime wh
func (b *Builder) WhereTime(column, operator string, values ...string) *Builder {

	value, pt := dealWhereValues(values...)

	return b.where("Time", column, operator, value, "and", pt)
}

// OrWhereTime ou
func (b *Builder) OrWhereTime(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Time", column, operator, value, "or", pt)
}

// WhereDay day
func (b *Builder) WhereDay(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Day", column, operator, value, "and", pt)
}

// OrWhereDay or
func (b *Builder) OrWhereDay(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Day", column, operator, value, "or", pt)
}

// WhereMonth month
func (b *Builder) WhereMonth(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Month", column, operator, value, "and", pt)
}

// OrWhereMonth or
func (b *Builder) OrWhereMonth(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Month", column, operator, value, "or", pt)
}

// WhereYear Add a "where year" statement to the query.
func (b *Builder) WhereYear(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Year", column, operator, value, "and", pt)
}

// OrWhereYear Add a "or where year" statement to the query.
func (b *Builder) OrWhereYear(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Year", column, operator, value, "or", pt)
}

// WhereDate data
func (b *Builder) WhereDate(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Date", column, operator, value, "and", pt)
}

// OrWhereDate or
func (b *Builder) OrWhereDate(column, operator string, values ...string) *Builder {
	value, pt := dealWhereValues(values...)

	return b.where("Date", column, operator, value, "or", pt)
}

func (b *Builder) addBindings(where map[string]string) {
	if len(where["value"]) > 0 {
		for _, value := range strings.Split(where["value"], ",") {
			b.Bindings.Wheres = append(b.Bindings.Wheres, strings.TrimSpace(value))
		}
	}
}

// Having clause to the query.
func (b *Builder) Having(column, operator string, values ...string) *Builder {

	var value string

	if len(values) > 0 {
		value = values[0]
	}

	return b.having(column, operator, value, "and")
}

// OrHaving clause to the query.
func (b *Builder) OrHaving(column, operator string, values ...string) *Builder {

	var value string

	if len(values) > 0 {
		value = values[0]
	}

	return b.having(column, operator, value, "or")
}

// Add a "having" clause to the query.
func (b *Builder) having(column, operator, value, logical string) *Builder {

	if invalidOperatorAndValue(operator, value) {
		// TODO
		// log.Fatal("Illegal operator and values combination.")
	}

	if value == "" {
		operator, value = "=", operator
	}

	having := map[string]string{
		"type":     "Basic",
		"column":   column,
		"operator": operator,
		"value":    value,
		"logical":  logical,
	}

	b.Components["havings"] = append(b.Components["havings"], having)

	b.addBindings(having)

	return b
}

// Join j
func (b *Builder) Join(table, first, operator string, second ...string) *Builder {

	return b.join(table, first, operator, dealValues(second...), "inner")
}

// RightJoin left join
func (b *Builder) RightJoin(table, first, operator string, second ...string) *Builder {
	return b.join(table, first, operator, dealValues(second...), "right")
}

// LeftJoin left join
func (b *Builder) LeftJoin(table, first, operator string, second ...string) *Builder {

	return b.join(table, first, operator, dealValues(second...), "left")
}

func (b *Builder) join(table, first, operator, second, joinType string) *Builder {
	if second == "" {
		operator, second = "=", operator
	}

	if b.invalidOperator(operator) {
		operator = "="
	}

	join := map[string]string{
		"type":     joinType,
		"table":    table,
		"first":    first,
		"operator": operator,
		"second":   second,
		"logical":  "on",
	}

	b.Components["joins"] = append(b.Components["joins"], join)

	return b
}

func (b *Builder) orWhereColumn(first, operator, second string) *Builder {
	return b.whereColumn(first, operator, second, "or")
}

// Add a "where" clause comparing two columns to the query.
func (b *Builder) whereColumn(first, operator, second, logical string) *Builder {

	if second == "" {
		operator, second = "=", operator
	}

	if b.invalidOperator(operator) {
		operator = "="
	}

	where := map[string]string{
		"type":     "Column",
		"first":    first,
		"operator": operator,
		"second":   second,
		"logical":  logical,
	}

	b.Components["wheres"] = append(b.Components["wheres"], where)

	return b
}

// WhereRaw where sql
func (b *Builder) WhereRaw(sql string, values ...string) *Builder {
	return b.whereRaw(sql, strings.Join(values, ","), "and")
}

// OrWhereRaw where sql
func (b *Builder) OrWhereRaw(sql string, values ...string) *Builder {
	return b.whereRaw(sql, strings.Join(values, ","), "or")
}

// Add a raw where clause to the query.
func (b *Builder) whereRaw(sql, value, logical string) *Builder {

	where := map[string]string{
		"type":    "Raw",
		"sql":     sql,
		"value":   value,
		"logical": logical,
	}

	b.Components["wheres"] = append(b.Components["wheres"], where)

	// Binding Values
	b.addBindings(where)
	return b
}

// Where Add a basic where clause to the query.
func (b *Builder) where(whereType, column, operator, value, logical, pt string) *Builder {

	if invalidOperatorAndValue(operator, value) {
		// TODO
		// log.Fatal("Illegal operator and values combination.")
	}

	if value == "" {
		operator, value = "=", operator
	}

	if b.invalidOperator(operator) {
		operator = "="
	}
	where := map[string]string{
		"type":     whereType,
		"column":   column,
		"operator": operator,
		"value":    value,
		"logical":  logical,
		"pt":       pt,
	}

	b.Components["wheres"] = append(b.Components["wheres"], where)

	// Binding Values
	b.addBindings(where)

	return b
}

// Where Add a basic where clause to the query.
func (b *Builder) Where(column, operator string, values ...string) *Builder {

	value, pt := dealWhereValues(values...)

	return b.where("Basic", column, operator, value, "and", pt)
}

// OrWhere Add an "or where" clause to the query.
func (b *Builder) OrWhere(column, operator string, values ...string) *Builder {

	value, pt := dealWhereValues(values...)

	return b.where("Basic", column, operator, value, "or", pt)
}

// Latest Add an "order by" clause for a timestamp to the query.
func (b *Builder) Latest(columns ...string) *Builder {

	if columns == nil {
		columns = append(columns, "created_at")
	}

	return b.OrderBy(columns[0], "desc")
}

// Oldest Add an "order by" clause for a timestamp to the query.
func (b *Builder) Oldest(columns ...string) *Builder {

	if columns == nil {
		columns = append(columns, "created_at")
	}

	return b.OrderBy(columns[0], "asc")
}

// Find Execute a query for a single record by ID.
func (b *Builder) Find(id int, columns ...string) map[string]interface{} {
	return b.where("Basic", "id", "=", itoa(id), "and", "").First(columns...)
}

// Value Get a single column's value from the first result of a query.
func (b *Builder) Value(column string) interface{} {
	return b.First(column)[column]
}

// First Execute the query and get the first result.
func (b *Builder) First(columns ...string) (result map[string]interface{}) {

	got := b.Take(1).Get(columns...)

	if len(got) > 0 {
		result = got[0]
	}

	return
}

// Get execute the query as a "select" statement.
func (b *Builder) Get(columns ...string) []map[string]interface{} {

	if columns == nil {
		columns = append(columns, "*")
	}

	b.Columns = columns
	b.Components["columns"] = nil

	results := b.RunSelect()

	return results
}

// Skip offset
func (b *Builder) Skip(n int) *Builder {
	b.Offset(n)
	return b
}

// Offset offset
func (b *Builder) Offset(n int) *Builder {
	if n < 0 {
		n = 0
	}

	if b.Unions != nil && len(b.Unions) > 0 {
		b.UnionOffsetNum = n
		b.Components["offset"] = nil
	} else {
		b.OffsetNum = n
		b.Components["offset"] = nil
	}

	return b
}

// ForPage Set the limit and offset for a given page.
// query.ForPage(5, 25)
func (b *Builder) ForPage(page int, perPages ...int) *Builder {

	perPage := 15
	if perPages != nil {
		perPage = perPages[0]
	}

	b.Skip((page - 1) * perPage).Take(perPage)
	return b
}

// Take Alias to set the "limit" value of the query.
func (b *Builder) Take(n int) *Builder {
	b.Limit(n)
	return b
}

// Limit limit
func (b *Builder) Limit(n int) *Builder {
	if n <= 0 {
		return b
	}

	if b.Unions != nil && len(b.Unions) > 0 {
		b.UnionLimitNum = n
		b.Components["limit"] = nil
	} else {
		b.LimitNum = n
		b.Components["limit"] = nil
	}

	return b
}

// Count Retrieve the "count" result of the query.
func (b *Builder) Count(column ...string) interface{} { return b.aggregate("count", column...) }

// Max Retrieve the maximum value of a given column.
func (b *Builder) Max(column string) interface{} { return b.aggregate("max", column) }

// Min Retrieve the minimum value of a given column.
func (b *Builder) Min(column string) interface{} { return b.aggregate("min", column) }

// Sum Retrieve the sum of the values of a given column.
func (b *Builder) Sum(column string) interface{} { return b.aggregate("sum", column) }

// Avg Retrieve the average of the values of a given column.
func (b *Builder) Avg(column string) interface{} { return b.aggregate("avg", column) }

// Average Retrieve the average of the values of a given column.
func (b *Builder) Average(column string) interface{} { return b.aggregate("avg", column) }

// Execute an aggregate function on the database.
func (b *Builder) aggregate(fn string, column ...string) interface{} {

	if column == nil {
		column = append(column, "*")
	}

	b.Aggregate = map[string]string{
		"column": column[0],
		"fn":     fn,
	}

	b.Components["aggregate"] = nil

	res := b.RunSelect()

	if res != nil {
		return res[0]["aggregate"]
	}

	return 0
}

// From Set the table which the query is targeting.
func (b *Builder) From(from string) *Builder {
	b.FromTable = from
	b.Components["from"] = nil

	return b
}

// Select Set the columns to be selected.
func (b *Builder) Select(columns ...string) *Builder {

	if columns == nil {
		columns = []string{"*"}
	}

	for _, value := range columns {
		b.Columns = append(b.Columns, value)
		b.Components["columns"] = nil
	}

	return b
}

// Distinct Force the query to only return distinct results.
func (b *Builder) Distinct() *Builder {
	b.IsDistinct = true
	return b
}
