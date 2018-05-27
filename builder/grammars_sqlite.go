package builder

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
