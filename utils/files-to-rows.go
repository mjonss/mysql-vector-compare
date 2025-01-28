// From ollama run deepseek-r1:70B
// > Please generate a program that takes all .md files in a directory and subdirectory and inserts it to a mysql database table with filename and content columns written in golang
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <directory> <mysql-connection-string>\n", os.Args[0])
	}

	dir := os.Args[1]
	connStr := os.Args[2]

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	defer db.Close()

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS markdown_files (
	    id bigint unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY,
		filename VARCHAR(255),
		content TEXT
	)`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Error creating table: %v\n", err)
	}

	stmt, err := db.PrepareContext(context.Background(), `INSERT INTO markdown_files (filename, content) VALUES (?, ?)`)
	if err != nil {
		log.Fatalf("Error preparing statement: %v\n", err)
	}
	defer stmt.Close()

	// Walk through directory and process .md files
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file: %v", err)
		}
		if !info.IsDir() && filepath.Ext(info.Name()) == ".md" {
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				return nil // continue processing other files
			}

			// Insert into database

			_, err = stmt.ExecContext(context.Background(), info.Name(), string(content))
			if err != nil {
				fmt.Printf("Error inserting file %s: %v\n", info.Name(), err)
				return nil
			}

			fmt.Printf("Inserted %s into database\n", info.Name())
		}
		return nil
	})
}
