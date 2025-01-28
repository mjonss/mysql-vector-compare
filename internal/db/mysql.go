package db

import (
	"database/sql"
	"fmt"
	"log"
	"unsafe"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDB struct {
	db   *sql.DB
	Name string
}

// NewMySQL creates a new MySQL-compatible database connection
// dbType can be "mysql", "mariadb", or "tidb"
func NewMySQL(dbType, host string, port int, user, password, dbname string, createDB bool) (*MySQLDB, error) {
	var dsn string
	if createDB {
		// TiDB needs to connect without database first to create it
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, "")
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, dbname)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s connection: %w", dbType, err)
	}

	if createDB {
		// Create and use database for TiDB
		if _, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname)); err != nil {
			return nil, fmt.Errorf("failed to create TiDB database: %w", err)
		}
		if _, err = db.Exec(fmt.Sprintf("USE %s", dbname)); err != nil {
			return nil, fmt.Errorf("failed to use TiDB database: %w", err)
		}
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping %s: %w", dbType, err)
	}

	return &MySQLDB{
		db:   db,
		Name: dbType,
	}, nil
}

func (m *MySQLDB) Query(q string) (*sql.Rows, error) {
	return m.db.Query(q)
}

func (m *MySQLDB) Exec(q string, args ...any) (sql.Result, error) {
	return m.db.Exec(q, args...)
}

func (m *MySQLDB) GetVersion() string {
	rs, err := m.db.Query("SELECT version()")
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	var version string
	for rs.Next() {
		var s string
		err = rs.Scan(&s)
		if err != nil {
			log.Fatalf("Failed to execute query: %v", err)
		}
		version += s
	}
	return version
}

func (m *MySQLDB) SupportColumn(colDef string) bool {
	_, err := m.db.Exec(fmt.Sprintf("create table vtWP (id int, v %s)", colDef))
	if err != nil {
		//log.Printf("Failed with column type 'VECTOR': %v", err)
		return false
	}
	_, err = m.db.Exec("drop table vtWP")
	if err != nil {
		log.Fatalf("Failed dropping test table vtWP: %v", err)
	}
	return true
}

func (m *MySQLDB) InsertVector(name string, vector []float32) error {
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(&vector[0])), len(vector)*4)
	log.Print(bytes)
	query := `INSERT INTO vt (v, note) VALUES (?, ?)`
	_, err := m.db.Exec(query, bytes, name)
	return err
}

func (m *MySQLDB) Close() error {
	return m.db.Close()
}
