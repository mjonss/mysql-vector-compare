package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type TiDB struct {
	db *sql.DB
}

func NewTiDB(host string, port int, user, password, dbname string) (*TiDB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, host, port, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &TiDB{db: db}, nil
}

func (t *TiDB) CreateVectorTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS vectors (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		name VARCHAR(255),
		vector JSON
	)`

	_, err := t.db.Exec(query)
	return err
}

func (t *TiDB) InsertVector(name string, vector []float32) error {
	query := `INSERT INTO vectors (name, vector) VALUES (?, ?)`
	_, err := t.db.Exec(query, name, vector)
	return err
}
