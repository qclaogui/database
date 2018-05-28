package builder

import (
	"strings"
)

// Grammars g
type Grammars struct {
	Builder        *Builder
	Prefix         string
	Placeholder    string
	PlaceholderNum int
}

// GetPlaceholder Get the query parameter place-holder for a value.
//	Parameter Placeholder Syntax
//	The syntax for placeholder parameters in prepared statements is database-specific. For example, comparing MySQL, PostgreSQL, and Oracle:
//	MySQL               PostgreSQL            Oracle
//	=====               ==========            ======
//	WHERE col = ?       WHERE col = $1        WHERE col = :col
//	VALUES(?, ?, ?)     VALUES($1, $2, $3)    VALUES(:val1, :val2, :val3)
func (g *Grammars) GetPlaceholder() string {
	if g.Placeholder != "?" {
		g.PlaceholderNum++
		return g.Placeholder + itoa(g.PlaceholderNum)
	}

	return "?"
}

// CompileInsert compile an insert statement into SQL.
func (g *Grammars) CompileInsert() {
	g.Builder.PSql = g.compileInsert()
	g.Builder.PArgs = g.Builder.Bindings.Wheres
}

// CompileDelete compile an delete statement into SQL.
func (g *Grammars) CompileDelete() {
	g.Builder.PSql = g.compileDelete()
	g.Builder.PArgs = g.Builder.Bindings.Wheres
}

// CompileUpdate compile an update statement into SQL.
func (g *Grammars) CompileUpdate() {
	g.Builder.PSql = g.compileUpdate()
	g.Builder.PArgs = g.Builder.Bindings.Wheres
}

// CompileSelect compile an select statement into SQL.
func (g *Grammars) CompileSelect() {
	g.Builder.PSql = g.compileSelect()
	g.Builder.PArgs = g.Builder.Bindings.Wheres
}

// CompileExists com
func (g *Grammars) CompileExists() {
	var sql strings.Builder
	sql.Grow(1024)
	sql.WriteString("select exists(")
	sql.WriteString(g.compileSelect())
	sql.WriteString(") as ")
	sql.WriteString(g.wrap("exists"))

	g.Builder.PSql = sql.String()
	g.Builder.PArgs = g.Builder.Bindings.Wheres
}

// SetTablePrefix Set the grammar's table prefix.
func (g *Grammars) SetTablePrefix(prefix string) {
	g.Prefix = prefix
}

// GetTablePrefix Get the grammar's table prefix.
func (g *Grammars) GetTablePrefix() string {
	return g.Prefix
}

// Wrap Get the grammar's table prefix.
func (g *Grammars) Wrap(value string) string {
	return g.wrap(value)
}

// Compile an insert statement into SQL.
func (g *Grammars) compileInsert() string {
	var sql strings.Builder
	sql.Grow(1024)
	sql.WriteString("insert into ")
	sql.WriteString(g.GetTablePrefix())
	sql.WriteString(g.Builder.FromTable)
	sql.WriteString("(")
	columns, parameters := make([]string, 0, len(g.Builder.Values[0])), make([]string, 0, len(g.Builder.Values))

	// map[string]string{ "name":  "Gopher", "email": "qclaogui@gmail.com"}
	for key := range g.Builder.Values[0] {
		columns = append(columns, key)
	}

	for _, m := range g.Builder.Values {
		parameters = append(parameters, "("+g.parameterize(columns, m)+")")
	}
	g.columnize(columns, &sql)
	sql.WriteString(") values ")
	sql.WriteString(strings.Join(parameters, ", "))

	return sql.String()
}

func (g *Grammars) compileDelete() string {
	var sql strings.Builder
	sql.Grow(1024)
	sql.WriteString("delete from ")
	sql.WriteString(g.GetTablePrefix())
	sql.WriteString(g.wrap(g.Builder.FromTable))

	g.compileComponentWheres(&sql)

	return sql.String()
}

// Compile an update statement into SQL.
func (g *Grammars) compileUpdate() string {

	var sql, updateColumns, joins strings.Builder
	updateColumns.Grow(1024)
	joins.Grow(1024)
	sql.Grow(1024)

	sql.WriteString("update ")
	sql.WriteString(g.GetTablePrefix())
	sql.WriteString(g.wrap(g.Builder.FromTable))

	if g.Builder.Components["joins"] != nil {
		g.compileComponentJoins(&joins)
	}

	sql.WriteString(joins.String())
	sql.WriteString(" set ")

	for key := range g.Builder.Values[0] {
		updateColumns.WriteString(g.wrap(key))
		updateColumns.WriteString(" = ")
		updateColumns.WriteString(g.GetPlaceholder())
		updateColumns.WriteString(", ")
	}

	sql.WriteString(strings.TrimRight(updateColumns.String(), ", "))
	g.compileComponentWheres(&sql)

	return sql.String()
}

func (g *Grammars) parameterize(columns []string, values map[string]string) string {
	var col strings.Builder
	col.Grow(1024)
	for k, v := range columns {
		g.Builder.Bindings.Wheres = append(g.Builder.Bindings.Wheres, values[v])
		if k > 0 {
			col.WriteString(", ")
			col.WriteString(g.GetPlaceholder())
		} else {
			col.WriteString(g.GetPlaceholder())
		}
	}

	return col.String()
}

// Compile a select query into SQL.
func (g *Grammars) compileSelect() string {
	var sql strings.Builder
	sql.Grow(1024)
	for _, component := range g.Builder.SelectComponents {
		if _, ok := g.Builder.Components[component]; ok {
			g.compileComponent(component, &sql)
		}
	}

	return sql.String()
}

// 		allocs := NumAllocs(func() {
// 		})
// 		fmt.Printf("\x1b[92m compileComponentColumns NumAllocs:\x1b[39m \x1b[91m %v \x1b[39m\n", allocs)
func (g *Grammars) compileComponent(component string, sql *strings.Builder) {
	switch component {
	case "aggregate":
		g.compileComponentAggregate(sql)
	case "columns":
		g.compileComponentColumns(sql)
	case "from":
		g.compileComponentFromTable(sql)
	case "joins":
		g.compileComponentJoins(sql)
	case "wheres":
		g.compileComponentWheres(sql)
	case "groups":
		g.compileComponentGroups(sql)
	case "havings":
		g.compileComponentHavings(sql)
	case "orders":
		g.compileComponentOrders(sql)
	case "limit":
		g.compileComponentLimitNum(sql)
	case "offset":
		g.compileComponentOffsetNum(sql)
	case "unions":
	case "lock":
	}
}

func (g *Grammars) compileComponentAggregate(sql *strings.Builder) {
	sql.WriteString("select ")
	sql.WriteString(g.Builder.Aggregate["fn"])
	sql.WriteString("(")
	if g.Builder.IsDistinct && g.Builder.Aggregate["column"] != "*" {
		sql.WriteString("distinct ")
		sql.WriteString(g.wrap(g.Builder.Aggregate["column"]))
	} else {
		sql.WriteString(g.wrap(g.Builder.Aggregate["column"]))
	}
	sql.WriteString(") as aggregate")
}

// Convert []string column names into a delimited string.
// Compile the "select *" portion of the query.
func (g *Grammars) compileComponentColumns(sql *strings.Builder) {
	if g.Builder.IsDistinct {
		sql.WriteString("select distinct ")
	} else {
		sql.WriteString("select ")
	}

	g.columnize(g.Builder.Columns, sql)
}

func (g *Grammars) compileComponentFromTable(sql *strings.Builder) {
	sql.WriteString(" from ")
	sql.WriteString(g.GetTablePrefix())
	sql.WriteString(g.Builder.FromTable)
}

func (g *Grammars) compileComponentJoins(sql *strings.Builder) {
	for _, join := range g.Builder.Components["joins"] {
		sql.WriteString(" ")
		sql.WriteString(join["type"])
		sql.WriteString(" join ")
		sql.WriteString(g.GetTablePrefix())
		sql.WriteString(join["table"])
		sql.WriteString(" ")
		sql.WriteString(join["logical"])
		sql.WriteString(" ")
		sql.WriteString(g.wrap(join["first"]))
		sql.WriteString(" ")
		sql.WriteString(join["operator"])
		sql.WriteString(" ")
		sql.WriteString(g.wrap(join["second"]))
	}
}

func (g *Grammars) compileComponentWheres(sql *strings.Builder) {

	sql.WriteString(" where ")
	for i, w := range g.Builder.Components["wheres"] {
		// skip first where logical
		if i > 0 {
			sql.WriteString(" ")
			sql.WriteString(w["logical"])
			sql.WriteString(" ")
		}

		if pt, ok := w["pt"]; ok {
			if pt == "(" {
				sql.WriteString("(")
			}
		}

		switch w["type"] {
		case "Basic":
			sql.WriteString(g.wrap(w["column"]))
			sql.WriteString(" ")
			sql.WriteString(w["operator"])
			sql.WriteString(" ")
			sql.WriteString(g.GetPlaceholder())
		case "Between":
			sql.WriteString(g.wrap(w["column"]))
			if v, ok := w["not"]; ok && v == "true" {
				sql.WriteString(" not between ? and ?")
			} else {
				sql.WriteString(" between ? and ?")
			}
		case "In":
			sql.WriteString(g.wrap(w["column"]))
			if v, ok := w["not"]; ok && v == "true" {
				sql.WriteString(" not in (")
			} else {
				sql.WriteString(" in (")
			}
			for i := range strings.Split(w["value"], ",") {
				if i == 0 {
					sql.WriteString(g.GetPlaceholder())
				} else {
					sql.WriteString(", ")
					sql.WriteString(g.GetPlaceholder())
				}
			}
			sql.WriteString(")")
		case "Date", "Year", "Month", "Day", "Time":
			sql.WriteString(strings.ToLower(w["type"]))
			sql.WriteString("(")
			sql.WriteString(g.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" ")
			sql.WriteString(g.GetPlaceholder())
		case "Column":
			sql.WriteString(g.wrap(w["first"]))
			sql.WriteString(" ")
			sql.WriteString(w["operator"])
			sql.WriteString(" ")
			sql.WriteString(g.wrap(w["second"]))
		case "Raw":
			sql.WriteString(w["sql"])
		default:
			panic("where type not Found")
		}

		if pt, ok := w["pt"]; ok {
			if pt == ")" {
				sql.WriteString(")")
			}
		}
	}
}
func (g *Grammars) compileComponentGroups(sql *strings.Builder) {
	sql.WriteString(" group by ")
	g.columnize(g.Builder.Groups, sql)
}

func (g *Grammars) compileComponentHavings(sql *strings.Builder) {

	sql.WriteString(" having ")
	for i, having := range g.Builder.Components["havings"] {
		if i > 0 {
			sql.WriteString(" ")
			sql.WriteString(having["logical"])
			sql.WriteString(" ")
		}
		if having["type"] == "Raw" {
			sql.WriteString(having["sql"])
		} else {
			sql.WriteString(g.wrap(having["column"]))
			sql.WriteString(" ")
			sql.WriteString(having["operator"])
			sql.WriteString(" ")
			sql.WriteString(g.GetPlaceholder())
		}
	}
}

func (g *Grammars) compileComponentOrders(sql *strings.Builder) {
	sql.WriteString(" order by ")
	for _, order := range g.Builder.Components["orders"] {
		sql.WriteString(g.wrap(order["column"]))
		sql.WriteString(" ")
		sql.WriteString(order["direction"])
	}
}

func (g *Grammars) compileComponentLimitNum(sql *strings.Builder) {
	sql.WriteString(" limit ")
	sql.WriteString(itoa(g.Builder.LimitNum))
}

func (g *Grammars) compileComponentOffsetNum(sql *strings.Builder) {
	sql.WriteString(" offset ")
	sql.WriteString(itoa(g.Builder.LimitNum))
}

func (g *Grammars) columnize(columns []string, sql *strings.Builder) {
	for i, value := range columns {
		if i > 0 {
			sql.WriteString(", ")
			sql.WriteString(g.wrap(value))
		} else {
			sql.WriteString(g.wrap(value))
		}
	}
}

// is a value has an alias. like: "email as user_email"
func hasAliasedAs(column string) (has bool) {
	for _, value := range strings.Fields(column) {
		if value == "as" {
			has = true
		}
	}
	return
}

func (g *Grammars) wrap(value string) string {
	var col strings.Builder
	col.Grow(100)

	if hasAliasedAs(value) {
		return g.wrapTwoValue(value, " as ", &col)
	}

	if strings.Contains(value, ".") {
		return g.wrapTwoValue(value, ".", &col)
	}
	col.Reset()

	return value
}

func (g *Grammars) wrapTwoValue(value, sep string, col *strings.Builder) (s string) {
	segments := strings.SplitN(value, sep, 2)
	col.WriteString(g.GetTablePrefix())
	col.WriteString(segments[0])
	col.WriteString(sep)
	if len(value) > 1 {
		col.WriteString(segments[1])
	}
	s = col.String()
	col.Reset()

	return
}
