package builder

// OracleGrammars Oracle
type OracleGrammars struct {
	Grammars
}

// SetBuilder set
func (og *OracleGrammars) SetBuilder(B *Builder) {
	og.Builder = B
}

// GetBuilder get
func (og *OracleGrammars) GetBuilder() *Builder {
	if og.Builder == nil {
		panic("GetBuilder PostgresGrammars error")
	}
	return og.Builder
}
