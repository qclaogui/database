package builder

// Connector c
type Connector interface {
	Connect()

	Insert() int64

	Update() int64

	// Select Run a select statement against the database.
	Select(bool) []map[string]interface{}

	Pretend(func()) []map[string]interface{}

	Delete() int64

	Table(string) *Builder

	Exists() bool

	// AffectingStatement Run an SQL statement and get the number of rows affected.
	AffectingStatement() int64
}
