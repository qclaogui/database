package builder

import (
	"io/ioutil"
	"log"
	"sync"

	"gopkg.in/yaml.v2"
)

var yDBConfig = []byte(`
default: mysql
sqlite:
   driver: sqlite3
   database: /absolute/path/to/gogogo.sqlite
   prefix:
mysql:
  driver: mysql
  read:
      host:
          - 
  write:
      host:
          - 
          - 127.0.0.1
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
  sslmode: disable`)

// DatabaseConfig use to load config
type DatabaseConfig struct {
	Default string       `yaml:"default"`
	Mysql   MysqlConfig  `yaml:"mysql"`
	Pgsql   PgsqlConfig  `yaml:"pgsql"`
	SQLite  SQLiteConfig `yaml:"sqlite"`
}

// SQLiteConfig sqlite
type SQLiteConfig struct {
	Driver   string `yaml:"driver"`
	Database string `yaml:"database"`
	Prefix   string `yaml:"prefix"`
}

// PgsqlConfig pgsql
type PgsqlConfig struct {
	Driver string `yaml:"driver"`
	Read   struct {
		Host []string `yaml:"host"`
	} `yaml:"read"`
	Write struct {
		Host []string `yaml:"host"`
	} `yaml:"write"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Charset  string `yaml:"charset"`
	Prefix   string `yaml:"prefix"`
	Sslmode  string `yaml:"sslmode"`
}

// MysqlConfig mysql
type MysqlConfig struct {
	Driver string `yaml:"driver"`
	Read   struct {
		Host []string `yaml:"host"`
	} `yaml:"read"`
	Write struct {
		Host []string `yaml:"host"`
	} `yaml:"write"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Database   string `yaml:"database"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Charset    string `yaml:"charset"`
	Prefix     string `yaml:"prefix"`
	Collation  string `yaml:"collation"`
	UnixSocket string `yaml:"unix_socket"`
}

// DBConfig config
type DBConfig struct {
	Driver    string
	ReadHost  []string
	WriteHost []string
	Host      string
	Port      string
	Database  string
	Username  string
	Password  string
	Charset   string
	Prefix    string
	// 	mysql
	Collation  string
	UnixSocket string
	// pgsql
	Sslmode string
}

func (dm *DatabaseManager) parseConfig(cName string) {
	switch cName {
	case "sqlite":
		dm.Config.Driver = dm.ymlConfig.SQLite.Driver
		dm.Config.Database = dm.ymlConfig.SQLite.Database
		dm.Config.Prefix = dm.ymlConfig.SQLite.Prefix
	case "mysql":
		dm.Config.Driver = dm.ymlConfig.Mysql.Driver
		dm.Config.ReadHost = dm.ymlConfig.Mysql.Read.Host
		dm.Config.WriteHost = dm.ymlConfig.Mysql.Write.Host
		dm.Config.Host = dm.ymlConfig.Mysql.Host
		dm.Config.Port = dm.ymlConfig.Mysql.Port
		dm.Config.Database = dm.ymlConfig.Mysql.Database
		dm.Config.Username = dm.ymlConfig.Mysql.Username
		dm.Config.Password = dm.ymlConfig.Mysql.Password
		dm.Config.Charset = dm.ymlConfig.Mysql.Charset
		dm.Config.Prefix = dm.ymlConfig.Mysql.Prefix
		dm.Config.Collation = dm.ymlConfig.Mysql.Collation
		dm.Config.UnixSocket = dm.ymlConfig.Mysql.UnixSocket
	case "pgsql":
		dm.Config.Driver = dm.ymlConfig.Pgsql.Driver
		dm.Config.ReadHost = dm.ymlConfig.Pgsql.Read.Host
		dm.Config.WriteHost = dm.ymlConfig.Pgsql.Write.Host
		dm.Config.Host = dm.ymlConfig.Pgsql.Host
		dm.Config.Port = dm.ymlConfig.Pgsql.Port
		dm.Config.Database = dm.ymlConfig.Pgsql.Database
		dm.Config.Username = dm.ymlConfig.Pgsql.Username
		dm.Config.Password = dm.ymlConfig.Pgsql.Password
		dm.Config.Charset = dm.ymlConfig.Pgsql.Charset
		dm.Config.Prefix = dm.ymlConfig.Pgsql.Prefix
		dm.Config.Sslmode = dm.ymlConfig.Pgsql.Sslmode
	}
	return
}

// DatabaseManager  database manager.
type DatabaseManager struct {
	once        sync.Once
	ymlConfig   DatabaseConfig
	ymlPath     string
	isLoaded    bool
	Config      DBConfig
	connections map[string]Connector
}

// Run return DBC(DB Connection) and DM(DatabaseManager)
func Run(ymlPath ...string) (Connector, *DatabaseManager) {

	dm := &DatabaseManager{}

	if ymlPath != nil && ymlPath[0] != "" {
		dm.ymlPath = ymlPath[0]
	}

	dm.once.Do(func() { dm.loadYmlConfig() })

	return dm.Connection(dm.ymlConfig.Default), dm
}

// load database
func (dm *DatabaseManager) loadYmlConfig() {
	if dm.isLoaded {
		return
	}

	if dm.ymlPath != "" {
		var err error
		yDBConfig, err = ioutil.ReadFile(dm.ymlPath)
		if err != nil {
			log.Fatalf("ReadFile err: #%v ", err)
		}
	}

	if err := yaml.Unmarshal(yDBConfig, &dm.ymlConfig); err != nil {
		log.Fatalf("yamlFile.Get err: #%v ", err)
	} else {
		dm.isLoaded = true
	}
}

// Connection Get a database connection instance.
// The name if you passed to the connection method should correspond to one of
// the listed(mysql,pgsql,sqlite) in yml file
// means one of MysqlConfig, PgsqlConfig, SQLiteConfig
func (dm *DatabaseManager) Connection(name string) Connector {

	if !supportedDrivers(name) {
		log.Fatalf("config name not support")
	}

	if !dm.hasConnection(name) {
		dm.makeConnection(name)
	}

	// log.Printf("\x1b[92m dm.connections[%#v]: %#v\x1b[39m", cName, dm.connections[cName])
	return dm.connections[name]
}

func supportedDrivers(name string) (support bool) {
	for _, v := range []string{"mysql", "pgsql", "sqlite"} {
		if name == v {
			support = true
		}
	}
	return
}

func (dm *DatabaseManager) hasConnection(name string) (exist bool) {
	if _, ok := dm.connections[name]; ok {
		exist = true
	}
	return
}

func (dm *DatabaseManager) makeConnection(name string) {

	dm.parseConfig(name)

	var conn Connector
	switch dm.Config.Driver {
	case "mysql":
		conn = NewMysqlConnection(dm.Config)
	case "postgres":
		conn = NewPostgresConnection(dm.Config)
	case "sqlite3":
		conn = NewSQLiteConnection(dm.Config)
	default:
		log.Fatalln("\x1b[31m DB Driver unknown \x1b[39m")
	}

	conn.Connect()

	dm.connections = map[string]Connector{name: conn}
}
