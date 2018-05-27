package builder

// OracleConnection sqlite
type OracleConnection struct {
	Connection
}

// NewOracleConnection TODO
func NewOracleConnection(c DBConfig) *OracleConnection {
	conn := &OracleConnection{Connection{
		Config: c,
		Grammar: &OracleGrammars{Grammars{
			Prefix: c.Prefix, Placeholder: ":"}}}}

	// 初始化builder
	conn.Grammar.SetBuilder(New(conn))

	return conn
}

// Table name
func (m *OracleConnection) Table(table string) *Builder { return m.Grammar.GetBuilder().From(table) }
