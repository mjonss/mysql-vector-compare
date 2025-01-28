package main

import (
	"log"
	"mysql-vector-compare/internal/db"
)

func main() {
	// Connect to MySQL
	mysqlDB, err := db.NewMySQL("localhost", 3306, "root", "password", "vectortest")
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	// Connect to MariaDB
	mariaDB, err := db.NewMariaDB("localhost", 3307, "root", "password", "vectortest")
	if err != nil {
		log.Fatalf("Failed to connect to MariaDB: %v", err)
	}

	// Connect to TiDB
	tiDB, err := db.NewTiDB("localhost", 4000, "root", "", "vectortest")
	if err != nil {
		log.Fatalf("Failed to connect to TiDB: %v", err)
	}

	// Create tables
	if err := mysqlDB.CreateVectorTable(); err != nil {
		log.Fatalf("Failed to create MySQL table: %v", err)
	}

	if err := mariaDB.CreateVectorTable(); err != nil {
		log.Fatalf("Failed to create MariaDB table: %v", err)
	}

	if err := tiDB.CreateVectorTable(); err != nil {
		log.Fatalf("Failed to create TiDB table: %v", err)
	}

	// Test vector insertion
	testVector := []float32{1.0, 2.0, 3.0, 4.0, 5.0}

	if err := mysqlDB.InsertVector("test_vector_mysql", testVector); err != nil {
		log.Printf("MySQL insertion error: %v", err)
	}

	if err := mariaDB.InsertVector("test_vector_mariadb", testVector); err != nil {
		log.Printf("MariaDB insertion error: %v", err)
	}

	if err := tiDB.InsertVector("test_vector_tidb", testVector); err != nil {
		log.Printf("TiDB insertion error: %v", err)
	}
}
