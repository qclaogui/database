package builder

// SQLiteConnection sqlite
type SQLiteConnection struct {
	Connection
}

// NewSQLiteConnection new
func NewSQLiteConnection(c DBConfig) *SQLiteConnection {
	conn := &SQLiteConnection{Connection{
		Config: c,
		Grammar: &SQLiteGrammars{Grammars{
			Prefix: c.Prefix, Placeholder: "?"}}}}

	// 初始化builder
	conn.Grammar.SetBuilder(New(conn))

	return conn
}

// Table name
func (m *SQLiteConnection) Table(table string) *Builder { return m.Grammar.GetBuilder().From(table) }
