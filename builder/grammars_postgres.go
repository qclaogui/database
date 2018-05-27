package builder

// PostgresGrammars g
type PostgresGrammars struct {
	Grammars
}

// SetBuilder set
func (m *PostgresGrammars) SetBuilder(B *Builder) {
	m.Builder = B
}

// GetBuilder get
func (m *PostgresGrammars) GetBuilder() *Builder {
	if m.Builder == nil {
		panic("GetBuilder PostgresGrammars error")
	}
	return m.Builder
}
