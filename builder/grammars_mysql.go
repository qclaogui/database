package builder

// MySqlGrammars mysql
type MySqlGrammars struct {
	Grammars
}

// SetBuilder set
func (m *MySqlGrammars) SetBuilder(B *Builder) {
	m.Builder = B
}

// GetBuilder get
func (m *MySqlGrammars) GetBuilder() *Builder {
	if m.Builder == nil {
		panic("GetBuilder MySqlGrammars error")
	}
	return m.Builder
}
