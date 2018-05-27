package builder

// PostgresConnection pgsql
type PostgresConnection struct {
	Connection
}

// NewPostgresConnection new
func NewPostgresConnection(c DBConfig) *PostgresConnection {
	conn := &PostgresConnection{Connection{
		Config: c,
		Grammar: &PostgresGrammars{Grammars{
			Prefix: c.Prefix, Placeholder: "$"}}}}

	// 初始化builder
	conn.Grammar.SetBuilder(New(conn))

	return conn
}

// Table name
func (m *PostgresConnection) Table(table string) *Builder { return m.Grammar.GetBuilder().From(table) }
