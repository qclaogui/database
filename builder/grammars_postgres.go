package builder

// PostgresGrammars g
type PostgresGrammars struct {
	Grammars
}

// SetBuilder set
func (pg *PostgresGrammars) SetBuilder(B *Builder) {
	pg.Builder = B
}

// GetBuilder get
func (pg *PostgresGrammars) GetBuilder() *Builder {
	if pg.Builder == nil {
		panic("GetBuilder PostgresGrammars error")
	}
	return pg.Builder
}
