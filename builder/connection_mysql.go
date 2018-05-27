package builder

import (
	_ "github.com/go-sql-driver/mysql"
)

// MysqlConnection c
type MysqlConnection struct {
	Connection
}

// NewMysqlConnection new
func NewMysqlConnection(c DBConfig) *MysqlConnection {

	conn := &MysqlConnection{Connection{
		Config: c,
		Grammar: &MySqlGrammars{Grammars{
			Prefix: c.Prefix, Placeholder: "?"}}}}

	// 初始化builder
	conn.Grammar.SetBuilder(New(conn))

	return conn
}

// Table name
func (m *MysqlConnection) Table(table string) *Builder { return m.Grammar.GetBuilder().From(table) }
