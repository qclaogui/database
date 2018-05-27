> This project is under development, If you want, Let's Go!

<div align="center">
  <h1>Database</h1>
</div>

- [Introduction](#introduction)
  - [Configuration](#configuration)
  - [Using Multiple Database Connections](#using-multiple-database-connections)
- [Retrieving Results](#retrieving-results)
- [Aggregates](#aggregates)
- [Selects](#selects)
- [Where Clauses](#where-clauses)
- [Ordering, Grouping, Limit, & Offset](#ordering-grouping-limit-and-offset)
- [Inserts](#inserts)
- [Updates](#updates)
- [Deletes](#deletes)

<a name="introduction"></a>
## Introduction
Database SQL builder, written in Go, provides a convenient to creating
and running database queries. Here is an [api server](examples/notes-api/app.go) example.

<a name="configuration"></a>
### Configuration

The database configuration is located at `database.yml`. In this file
you may define all of your database connections, as well as specify
which connection should be used by default. Examples for most of the
supported database systems are provided in this file.

```yml
default: mysql
sqlite:
   driver: sqlite3
   database: /absolute/path/to/gogogo.sqlite
   prefix:
mysql:
  driver: mysql
  host: localhost
  port: 3306
  database: gogogo
  username: root
  password:
  unix_socket:
  charset: utf8mb4
  collation: utf8mb4_unicode_ci
  prefix:
pgsql:
  driver: postgres
  host: 127.0.0.1
  port: 5432
  database: gogogo
  username: qclaogui
  password:
  charset: utf8
  prefix:
  sslmode: disable
```

<a name="using-multiple-database-connections"></a>
### Using Multiple Database Connections

When using multiple connections, you may access each connection via the
`connection` method on the `DM`. The `name` passed to the `connection`
method should correspond to one of the connections listed in your
`database.yml` configuration file:
```go
package main

import "github.com/qclaogui/database/builder"

DB, DM := builder.Run("/absolute/path/to/database.yml")

users :=DB.Table("users").Get()
// or
users = DM.Connection("pgsql").Table("users").Get()
```

<a name="retrieving-results"></a>
## Retrieving Results

#### Retrieving All Rows From A Table

You may use the Table method on the DB(Connector interface) to begin a
query
```go
DB, _ := builder.Run("/absolute/path/to/database.yml")

users := DB.Table("users").Get()
```
#### Retrieving A Single Row / Column From A Table

```go
users := DB.Table("users").Limit(1).Get()
```
Value method will return the value of the column directly:

```go
users := DB.Table("users").Where("name", "John").Value("email")
```

<a name="aggregates"></a>
### Aggregates

The query builder also provides a variety of aggregate methods such as `count`, `max`, `min`, `avg`, and `sum`.
You may call any of these methods after constructing your query:

```go
users := DB.Table("users").Count()

users := DB.Table("users").Max("id")

price := DB.Table("orders").Where("finalized", "1").Avg("price")
```

<a name="selects"></a>
## Selects

#### Specifying A Select Clause
Using the `Select` method, you can specify a custom `Select` clause for
the query:
```go
users :=DB.Table("users").Select("name", "email as user_email").Get()

```
The `Distinct` method allows you to force the query to return distinct
results:
```go
users :=DB.Table("users").Distinct().Get()

```

<a name="where-clauses"></a>
## Where Clauses

#### Simple Where Clauses
The most basic call to `where` requires three arguments. The first argument is the name of the column. The second argument is an operator, which can be any of the database's supported operators. Finally, the third argument is the value to evaluate against the column.
You may use the `where` method

```go
users :=DB.Table("users").Where("votes", "=", "100").Get()
// or 
users :=DB.Table("users").Where("votes", "100").Get()

```
you may use a variety of other operators when writing a `where` clause:
```go
users :=DB.Table("users").Where("votes",">=", "100").Get()

users :=DB.Table("users").Where("votes","<>", "100").Get()

users :=DB.Table("users").Where("votes","like", "T%").Get()

```

**whereBetween**

The `whereBetween` method verifies:

```go
users :=DB.Table("users").WhereBetween("created_at", "2017-01-08", "2018-03-06").Get()
```
**whereBetween / WhereNotBetween**

The `whereBetween` method verifies that a column's value is between two values:

```go
users :=DB.Table("users").WhereBetween("votes", "1", "100").Get()

users :=DB.Table("users").WhereNotBetween("votes", "1", "100").Get()
```

**WhereIn / whereNotIn**

The `WhereIn` method verifies that a given column's value is contained within the given array:
```go
users :=DB.Table("users").WhereIn("id", "1", "2","3").Get()
// or 
users :=DB.Table("users").WhereIn("id", []string{"1","2","3"}...).Get()

users :=DB.Table("users").WhereNotIn("id", "1", "2","3").Get()

users :=DB.Table("users").WhereNotIn("id", []string{"1","2","3"}...).Get()

```

**WhereDate / WhereMonth / WhereDay / WhereYear / WhereTime**
The `WhereDate` method may be used to compare a column's value against a date:

```go
users :=DB.Table("users").WhereDate("created_at", "2018-05-20").Get()
```
The `WhereMonth` method may be used to compare a column's value against a
specific month of a year:
```go
users :=DB.Table("users").WhereMonth("created_at", "12").Get()

```

The `WhereDay` method may be used to compare a column's value against a
specific day of a month:

```go
users :=DB.Table("users").WhereDay("created_at", "12").Get()

```
The `WhereYear` method may be used to compare a column's value against a
specific year:

```go
users :=DB.Table("users").WhereYear("created_at", "2018").Get()

```
The `WhereTime` method may be used to compare a column's value against a
specific time:

```go
users :=DB.Table("users").WhereTime("created_at","=", "12:30:15").Get()

```

<a name="ordering-grouping-limit-and-offset"></a>
## Ordering, Grouping, Limit, & Offset

#### OrderBy

The `OrderBy` method allows you to sort the result of the query by a given column. The first argument to the `OrderBy` method should be the column you wish to sort by, while the second argument controls the direction of the sort and may be either `asc` or `desc`:
```go
users :=DB.Table("users").OrderBy("name", "desc").Get()
```

#### GroupBy / Having

The `GroupBy` and `Having` methods may be used to group the query results. The `Having` method's signature is similar to that of the `Where` method:
```go
users :=DB.Table("users").GroupBy("account_id").Having("account_id", ">", "10").Get()
```

#### Skip / Take
To limit the number of results returned from the query, or to skip a given number of results in the query, you may use the `Skip` and `Take` methods:
```go
users :=DB.Table("users").Skip(10).Take(5).Get()

```
Alternatively, you may use the `limit` and `offset` methods:
```go
users :=DB.Table("users").Skip(10).Limit(5).Get()

```

<a name="inserts"></a>
## Inserts

The query builder also provides an `Insert` method for inserting records
into the database table.:
```go
var usersData = []map[string]string{map[string]string{
			"name":  "gopher1",
			"email": "gopher1@qq.com",
		}, map[string]string{
			"name":  "gopher2",
			"email": "gopher2@qq.com",
		}}

users :=DB.Table("users").Insert(usersData)
```
<a name="updates"></a>
## Updates
Of course, the query builder can also update existing records using the
`Update` method
```go
var updateData =  map[string]string{
			"name":  "gopher2",
			"email": "gopher2@gmail.com",
		}

users :=DB.Table("users").Update(updateData)
```


<a name="deletes"></a>
## Deletes

The query builder may also be used to delete records from the table via
the `Delete` method. You may constrain `delete` statements by adding
`Where` clauses before calling the `Delete` method:

```go
users :=DB.Table("users").Delete()

users :=DB.Table("users").Where('votes', '>', "100").Delete()
```