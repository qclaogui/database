package builder

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Connection default DB connection
type Connection struct {
	DB                *sql.DB
	DBRead            *sql.DB                  // The DB for read.
	Config            DBConfig                 // The database connection configuration options.
	Grammar           Grammar                  // The query grammar implementation.
	queryLog          []map[string]interface{} // All of the queries run against the connection.
	loggingQueries    bool                     // Indicates whether queries are being logged.
	recordsIsModified bool                     // Indicates if changes have been made to the database.
}

func hasReadWrite(c *DBConfig) (hasRead bool) {
	if c.ReadHost != nil && c.ReadHost[0] != "" &&
		c.WriteHost != nil && c.WriteHost[0] != "" {
		hasRead = true
	}
	return
}

func makeDB(config DBConfig, DBType ...string) *sql.DB {

	if DBType != nil && DBType[0] == "write" {
		config.Host = config.WriteHost[0]
		// log.Printf("\x1b[92m DB Write Host: %#v \x1b[39m", config.Host)
	}
	if DBType != nil && DBType[0] == "read" {
		config.Host = config.ReadHost[0]
		// log.Printf("\x1b[92m DB Read Host: %#v \x1b[39m", config.Host)
	}

	dsn := getDsn(config)
	// log.Printf("\x1b[92m sql.Open(%#v, %#v) \x1b[39m", config.Driver, dsn)
	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		log.Fatalf("\x1b[31m sql.Open:\x1b[39m %s", err.Error())
	}
	// 建立连接
	if err = db.Ping(); err != nil {
		log.Fatalf("\x1b[31m sql.Ping: \x1b[39m %s", err.Error())
	}

	return db
}

// configureDBDsn
func configureDBDsn(config DBConfig) (params string) {
	switch config.Driver {
	case "mysql":
		params = "loc=Local"
		if config.Collation != "" {
			params += "&collation=" + config.Collation
		}
		if config.Charset != "" {
			params += "&charset=" + config.Charset
		}
		return
	case "postgres":
		port := "5432"
		if config.Port != "" {
			port = config.Port
		}
		params = "port=" + port
		if config.Sslmode != "" {
			params += "&sslmode=" + config.Sslmode
		}
	case "sqlite3":
		return
	default:
		return
	}
	return
}

// getDsn create a DSN string from a configuration.
// like:
// user@unix(/path/to/socket)/dbname
// root:pw@unix(/tmp/mysql.sock)/myDatabase?loc=Local

// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// username:password@protocol(address)/dbname?param=value
// user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
// user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname?timeout=90s&collation=utf8mb4_unicode_ci&charset=utf8mb4
// dsn := "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
func getDsn(config DBConfig) string {
	var dsn string

	if config.UnixSocket != "" {
		dsn = getSocketDsn(config)
	} else {
		dsn = getHostDsn(config)
	}

	return dsn + "?" + configureDBDsn(config)
}

// getSocketDsn get the DSN string for a socket configuration.
// user@unix(/path/to/socket)/dbname
// root:pw@unix(/tmp/mysql.sock)/myDatabase?loc=Local
func getSocketDsn(config DBConfig) string {
	return config.Username + ":" + config.Password + "unix(" + config.UnixSocket + ")/" + config.Database
}

// getHostDsn get the DSN string for a host / port configuration.
// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
// username:password@protocol(address)/dbname?param=value
// user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true
// user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname?timeout=90s&collation=utf8mb4_unicode_ci&charset=utf8mb4
// connStr := "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
// db, err := sql.Open("postgres", connStr)
func getHostDsn(config DBConfig) string {
	switch config.Driver {
	case "mysql":
		return config.Username + ":" + config.Password + "@tcp(" + config.Host + ":" + config.Port + ")/" + config.Database
	case "postgres":
		return "postgres://" + config.Username + ":" + config.Password + "@" + config.Host + "/" + config.Database
	case "sqlite3":
		return config.Database
	default:
		return config.Database
	}
}

// Connect Establish a connection based on the configuration.
func (m *Connection) Connect() {

	if hasReadWrite(&m.Config) {
		// 1. create Write Connection
		m.DB = makeDB(m.Config, "write")
		// 2. create Read Connection
		m.DBRead = makeDB(m.Config, "read")
	} else {
		m.DB = makeDB(m.Config)
	}
}

// Run a SQL statement and log its execution context.
func (m *Connection) run(callback func() ([]map[string]interface{}, int64, error)) ([]map[string]interface{}, int64) {

	start := time.Now()

	// resets the Builder
	m.Grammar.GetBuilder().Reset()

	// 开始执行callback 返回结果集，受影响的行数，发生错误
	result, rowCnt, err := callback()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	if m.Grammar.GetBuilder().debug {
		log.Printf("\x1b[92m DEBUG SQL:\x1b[39m %#v\n\t\x1b[92m Bindings:\x1b[39m %v Use: %v\n",
			m.Grammar.GetBuilder().PSql, m.Grammar.GetBuilder().PArgs, time.Since(start))
	}

	// Once we have run the query we will calculate the time that it took to run and
	// then log the query
	// m.logQuery("LOG: ", m.Grammar.GetBuilder().PSql, m.Grammar.GetBuilder().PArgs, time.Since(start))

	return result, rowCnt
}

// Log a query in the connection's query log.
func (m *Connection) logQuery(action, query string, bindings []interface{}, elapsed time.Duration) {
	if m.loggingQueries {
		m.queryLog = append(m.queryLog, map[string]interface{}{
			"action":   action,
			"query":    query,
			"bindings": bindings,
			"time":     elapsed,
		})
	}
}

// AffectingStatement Run an SQL statement and get the number of rows affected.
func (m *Connection) AffectingStatement() int64 {

	_, rowCnt := m.run(func() ([]map[string]interface{}, int64, error) {

		stmt, err := m.DB.Prepare(m.Grammar.GetBuilder().PSql)
		if err != nil {
			return nil, 0, &queryError{PSql: m.Grammar.GetBuilder().PSql, PArgs: m.Grammar.GetBuilder().PArgs, Err: err}
		}
		defer stmt.Close()

		res, err := stmt.Exec(m.Grammar.GetBuilder().PArgs...)
		if err != nil {
			return nil, 0, &queryError{PSql: m.Grammar.GetBuilder().PSql, PArgs: m.Grammar.GetBuilder().PArgs, Err: err}
		}
		// lastId, _ := res.LastInsertId()
		rowCnt, _ := res.RowsAffected()

		m.recordsHaveBeenModified(rowCnt > 0)

		return nil, rowCnt, nil
	})

	return rowCnt
}

// Select Run a select statement against the database.
func (m *Connection) Select(useReadDB bool) []map[string]interface{} {

	// compile an select statement into SQL.
	m.Grammar.CompileSelect()

	results, _ := m.run(func() ([]map[string]interface{}, int64, error) {
		var stmt *sql.Stmt
		var err error
		if useReadDB && m.DBRead != nil {
			log.Printf("\x1b[92m Select use m.DBRead: \x1b[39m%#v", m.DBRead)
			stmt, err = m.DBRead.Prepare(m.Grammar.GetBuilder().PSql)
		} else {
			// log.Printf("\x1b[92m Select use m.DB: \x1b[39m%#v", m.DB)
			stmt, err = m.DB.Prepare(m.Grammar.GetBuilder().PSql)
		}

		if err != nil {
			return nil, 0, &queryError{PSql: m.Grammar.GetBuilder().PSql, PArgs: m.Grammar.GetBuilder().PArgs, Err: err}
		}
		defer stmt.Close()

		rows, err := stmt.Query(m.Grammar.GetBuilder().PArgs...)
		if err != nil {
			return nil, 0, &queryError{PSql: m.Grammar.GetBuilder().PSql, PArgs: m.Grammar.GetBuilder().PArgs, Err: err}
		}
		defer rows.Close()

		columns, _ := rows.Columns()

		scanArgs, values := make([]interface{}, len(columns)), make([]sql.RawBytes, len(columns))

		for key := range values {
			scanArgs[key] = &values[key]
		}
		rowsMap := make([]map[string]interface{}, 0, 10)

		for rows.Next() {
			err := rows.Scan(scanArgs...)
			if err != nil {
				return nil, 0, &queryError{PSql: m.Grammar.GetBuilder().PSql, PArgs: m.Grammar.GetBuilder().PArgs, Err: err}
			}

			rowMap := make(map[string]interface{})
			for key, value := range values {
				rowMap[columns[key]] = string(value)
			}
			rowsMap = append(rowsMap, rowMap)
		}

		if err = rows.Err(); err != nil {
			return nil, 0, &queryError{PSql: m.Grammar.GetBuilder().PSql, PArgs: m.Grammar.GetBuilder().PArgs, Err: err}
		}

		return rowsMap, 0, nil
	})

	return results
}

// Exists a select statement
func (m *Connection) Exists() bool {

	// compile an select statement into SQL.
	m.Grammar.CompileExists()

	if m.Grammar.GetBuilder().debug {
		log.Fatalf("DEBUG SQL:\033[0;32m%v\033[0m; bindings: \033[0;32m %v\033[0m",
			m.Grammar.GetBuilder().PSql, m.Grammar.GetBuilder().PArgs)
	}
	// TODO
	m.Grammar.GetBuilder().Reset()

	return true
}

func (m *Connection) recordsHaveBeenModified(bo bool) {
	if !m.recordsIsModified {
		m.recordsIsModified = bo
	}
}

// Insert Run an insert statement against the database.
func (m *Connection) Insert() int64 {

	m.Grammar.CompileInsert()

	return m.AffectingStatement()
}

// Update Run an update statement against the database.
func (m *Connection) Update() int64 {

	m.Grammar.CompileUpdate()

	return m.AffectingStatement()
}

// Delete Run an delete statement against the database.
func (m *Connection) Delete() int64 {

	m.Grammar.CompileDelete()

	return m.AffectingStatement()
}

func (m *Connection) execInTransaction(B *Builder) {

	tx, err := m.DB.Begin()
	if err != nil {
		log.Fatalf("Whoops! Transaction Begin():%v", err.Error())
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(B.PSql)
	if err != nil {
		log.Fatalf("Whoops! Transaction Prepare():%v", err.Error())
	}

	_, err = stmt.Exec(B.PArgs...)
	if err != nil {
		log.Fatalf("Whoops! Transaction Exec():%v", err.Error())
	}

	// for i := 0; i < 10; i++ {
	//	_, err = stmt.Exec(i)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	// }

	err = tx.Commit()
	if err != nil {
		log.Fatalf("Whoops! Transaction Commit():%v", err.Error())
	}
	stmt.Close()
}
