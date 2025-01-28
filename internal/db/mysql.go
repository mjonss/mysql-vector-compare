package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDB struct {
	db *sql.DB
}

func NewMySQL(host string, port int, user, password, dbname string) (*MySQLDB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &MySQLDB{db: db}, nil
}

func (m *MySQLDB) CreateVectorTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS vectors (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255),
		vector JSON
	)`

	_, err := m.db.Exec(query)
	return err
}

func (m *MySQLDB) InsertVector(name string, vector []float32) error {
	query := `INSERT INTO vectors (name, vector) VALUES (?, ?)`
	_, err := m.db.Exec(query, name, vector)
	return err
}
