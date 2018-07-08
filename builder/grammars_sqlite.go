package builder

import (
	"strings"
)

// SQLiteGrammars SQLite
type SQLiteGrammars struct {
	Grammars
}

// SetBuilder set
func (sg *SQLiteGrammars) SetBuilder(B *Builder) {
	sg.Builder = B
}

// GetBuilder get
func (sg *SQLiteGrammars) GetBuilder() *Builder {
	if sg.Builder == nil {
		panic("GetBuilder PostgresGrammars error")
	}
	return sg.Builder
}

// CompileSelect compile an select statement into SQL.
func (sg *SQLiteGrammars) CompileSelect() { sg.Builder.PSql = sg.compileSelect() }

// CompileSelect compile an select statement into SQL.
func (sg *SQLiteGrammars) compileSelect() string {
	var sql strings.Builder
	sql.Grow(1024)
	for _, component := range sg.Builder.SelectComponents {
		if _, ok := sg.Builder.Components[component]; ok {
			sg.compileComponent(component, &sql)
		}
	}

	return sql.String()
}

func (sg *SQLiteGrammars) compileComponent(component string, sql *strings.Builder) {
	switch component {
	case "aggregate":
		sg.compileComponentAggregate(sql)
	case "columns":
		sg.compileComponentColumns(sql)
	case "from":
		sg.compileComponentFromTable(sql)
	case "joins":
		sg.compileComponentJoins(sql)
	case "wheres":
		sg.compileComponentWheres(sql)
	case "groups":
		sg.compileComponentGroups(sql)
	case "havings":
		sg.compileComponentHavings(sql)
	case "orders":
		sg.compileComponentOrders(sql)
	case "limit":
		sg.compileComponentLimitNum(sql)
	case "offset":
		sg.compileComponentOffsetNum(sql)
	case "unions":
	case "lock":
	}
}

func (sg *SQLiteGrammars) compileComponentWheres(sql *strings.Builder) {
	sql.WriteString(" where ")
	for i, w := range sg.Builder.Components["wheres"] {
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
			sql.WriteString(sg.wrap(w["column"]))
			sql.WriteString(" ")
			sql.WriteString(w["operator"])
			sql.WriteString(" ")
			sql.WriteString(sg.GetPlaceholder(w["value"]))
		case "Between":
			sql.WriteString(sg.wrap(w["column"]))
			p := sg.GetPlaceholder(w["value"])
			if v, ok := w["not"]; ok && v == "true" {
				sql.WriteString(" not between ")
				sql.WriteString(p)
				sql.WriteString(" and ")
				sql.WriteString(p)
			} else {
				sql.WriteString(" between ")
				sql.WriteString(p)
				sql.WriteString(" and ")
				sql.WriteString(p)
			}
		case "In":
			sql.WriteString(sg.wrap(w["column"]))
			if v, ok := w["not"]; ok && v == "true" {
				sql.WriteString(" not in (")
			} else {
				sql.WriteString(" in (")
			}
			for i := range strings.Split(w["value"], ",") {
				if i == 0 {
					sql.WriteString(sg.GetPlaceholder(w["value"]))
				} else {
					sql.WriteString(", ")
					sql.WriteString(sg.GetPlaceholder(w["value"]))
				}
			}
			sql.WriteString(")")
		case "Year":
			sql.WriteString("strftime('%Y', ")
			sql.WriteString(sg.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(sg.GetPlaceholder(w["value"]))
			sql.WriteString(" as text)")
		case "Month":
			sql.WriteString("strftime('%m', ")
			sql.WriteString(sg.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(sg.GetPlaceholder(w["value"]))
			sql.WriteString(" as text)")
		case "Date":
			sql.WriteString("strftime('%Y-%m-%d', ")
			sql.WriteString(sg.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(sg.GetPlaceholder(w["value"]))
			sql.WriteString(" as text)")
		case "Day":
			sql.WriteString("strftime('%d', ")
			sql.WriteString(sg.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(sg.GetPlaceholder(w["value"]))
			sql.WriteString(" as text)")
		case "Time":
			sql.WriteString("strftime('%H:%M:%S', ")
			sql.WriteString(sg.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(sg.GetPlaceholder(w["value"]))
			sql.WriteString(" as text)")
		case "Column":
			sql.WriteString(sg.wrap(w["first"]))
			sql.WriteString(" ")
			sql.WriteString(w["operator"])
			sql.WriteString(" ")
			sql.WriteString(sg.wrap(w["second"]))
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
