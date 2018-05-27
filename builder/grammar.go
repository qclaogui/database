package builder

// Grammar Sql
type Grammar interface {
	// SetTablePrefix Set the grammar's table prefix.
	SetTablePrefix(prefix string)

	// GetTablePrefix Get the grammar's table prefix.
	GetTablePrefix() string

	SetBuilder(*Builder)

	GetBuilder() *Builder

	CompileInsert()

	CompileDelete()

	CompileUpdate()

	CompileSelect()

	CompileExists()

	Wrap(string) string
}
