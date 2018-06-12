package builder

import (
	"strings"
)

// SQLiteGrammars SQLite
type SQLiteGrammars struct {
	Grammars
}

// SetBuilder set
func (m *SQLiteGrammars) SetBuilder(B *Builder) {
	m.Builder = B
}

// GetBuilder get
func (m *SQLiteGrammars) GetBuilder() *Builder {
	if m.Builder == nil {
		panic("GetBuilder PostgresGrammars error")
	}
	return m.Builder
}

// CompileSelect compile an select statement into SQL.
func (m *SQLiteGrammars) CompileSelect() {
	m.Builder.PSql = m.compileSelect()
	m.Builder.PArgs = m.Builder.Bindings.Wheres
}

// CompileSelect compile an select statement into SQL.
func (m *SQLiteGrammars) compileSelect() string {
	var sql strings.Builder
	sql.Grow(1024)
	for _, component := range m.Builder.SelectComponents {
		if _, ok := m.Builder.Components[component]; ok {
			m.compileComponent(component, &sql)
		}
	}

	return sql.String()
}

func (m *SQLiteGrammars) compileComponent(component string, sql *strings.Builder) {
	switch component {
	case "aggregate":
		m.compileComponentAggregate(sql)
	case "columns":
		m.compileComponentColumns(sql)
	case "from":
		m.compileComponentFromTable(sql)
	case "joins":
		m.compileComponentJoins(sql)
	case "wheres":
		m.compileComponentWheres(sql)
	case "groups":
		m.compileComponentGroups(sql)
	case "havings":
		m.compileComponentHavings(sql)
	case "orders":
		m.compileComponentOrders(sql)
	case "limit":
		m.compileComponentLimitNum(sql)
	case "offset":
		m.compileComponentOffsetNum(sql)
	case "unions":
	case "lock":
	}
}

func (m *SQLiteGrammars) compileComponentWheres(sql *strings.Builder) {
	sql.WriteString(" where ")
	for i, w := range m.Builder.Components["wheres"] {
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
			sql.WriteString(m.wrap(w["column"]))
			sql.WriteString(" ")
			sql.WriteString(w["operator"])
			sql.WriteString(" ")
			sql.WriteString(m.GetPlaceholder())
		case "Between":
			sql.WriteString(m.wrap(w["column"]))
			if v, ok := w["not"]; ok && v == "true" {
				sql.WriteString(" not between ? and ?")
			} else {
				sql.WriteString(" between ? and ?")
			}
		case "In":
			sql.WriteString(m.wrap(w["column"]))
			if v, ok := w["not"]; ok && v == "true" {
				sql.WriteString(" not in (")
			} else {
				sql.WriteString(" in (")
			}
			for i := range strings.Split(w["value"], ",") {
				if i == 0 {
					sql.WriteString(m.GetPlaceholder())
				} else {
					sql.WriteString(", ")
					sql.WriteString(m.GetPlaceholder())
				}
			}
			sql.WriteString(")")
		case "Year":
			sql.WriteString("strftime('%Y', ")
			sql.WriteString(m.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(m.GetPlaceholder())
			sql.WriteString(" as text)")
		case "Month":
			sql.WriteString("strftime('%m', ")
			sql.WriteString(m.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(m.GetPlaceholder())
			sql.WriteString(" as text)")
		case "Date":
			sql.WriteString("strftime('%Y-%m-%d', ")
			sql.WriteString(m.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(m.GetPlaceholder())
			sql.WriteString(" as text)")
		case "Day":
			sql.WriteString("strftime('%d', ")
			sql.WriteString(m.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(m.GetPlaceholder())
			sql.WriteString(" as text)")
		case "Time":
			sql.WriteString("strftime('%H:%M:%S', ")
			sql.WriteString(m.wrap(w["column"]))
			sql.WriteString(") ")
			sql.WriteString(w["operator"])
			sql.WriteString(" cast(")
			sql.WriteString(m.GetPlaceholder())
			sql.WriteString(" as text)")
		case "Column":
			sql.WriteString(m.wrap(w["first"]))
			sql.WriteString(" ")
			sql.WriteString(w["operator"])
			sql.WriteString(" ")
			sql.WriteString(m.wrap(w["second"]))
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
