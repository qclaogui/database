package builder

// MySqlGrammars mysql
type MySqlGrammars struct {
	Grammars
}

// SetBuilder set
func (mg *MySqlGrammars) SetBuilder(B *Builder) {
	mg.Builder = B
}

// GetBuilder get
func (mg *MySqlGrammars) GetBuilder() *Builder {
	if mg.Builder == nil {
		panic("GetBuilder MySqlGrammars error")
	}
	return mg.Builder
}
