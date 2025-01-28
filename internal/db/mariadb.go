package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type MariaDB struct {
	db *sql.DB
}

func NewMariaDB(host string, port int, user, password, dbname string) (*MariaDB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &MariaDB{db: db}, nil
}

func (m *MariaDB) CreateVectorTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS vectors (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255),
		vector JSON
	)`

	_, err := m.db.Exec(query)
	return err
}

func (m *MariaDB) InsertVector(name string, vector []float32) error {
	query := `INSERT INTO vectors (name, vector) VALUES (?, ?)`
	_, err := m.db.Exec(query, name, vector)
	return err
}
