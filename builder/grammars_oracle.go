package builder

// OracleGrammars Oracle
type OracleGrammars struct {
	Grammars
}

// SetBuilder set
func (m *OracleGrammars) SetBuilder(B *Builder) {
	m.Builder = B
}

// GetBuilder get
func (m *OracleGrammars) GetBuilder() *Builder {
	if m.Builder == nil {
		panic("GetBuilder PostgresGrammars error")
	}
	return m.Builder
}
