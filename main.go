package main

import (
	"fmt"
	"log"
	"mysql-vector-compare/internal/db"
	"strings"
	"unsafe"
)

func main() {
	// Connect to MySQL
	mysqlDB, err := db.NewMySQL("mysql", "localhost", 13306, "root", "password", "vectortest", false)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlDB.Close()

	myvecDB, err := db.NewMySQL("myvec", "localhost", 13000, "root", "", "vectortest", true)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer myvecDB.Close()

	mariaDB, err := db.NewMySQL("mariadb", "localhost", 13307, "root", "password", "vectortest", false)
	if err != nil {
		log.Fatalf("Failed to connect to MariaDB: %v", err)
	}
	defer mariaDB.Close()

	tiDB, err := db.NewMySQL("tidb", "localhost", 14000, "root", "", "vectortest", true)
	if err != nil {
		log.Fatalf("Failed to connect to TiDB: %v", err)
	}
	defer tiDB.Close()

	fmt.Printf("MySQL Version: %s\n", mysqlDB.GetVersion())
	fmt.Printf("MyVECTOR Version: %s\n", myvecDB.GetVersion())
	fmt.Printf("TiDB Version: %s\n", tiDB.GetVersion())
	fmt.Printf("MariaDB Version: %s\n", mariaDB.GetVersion())

	fmt.Printf("Feature\tMySQL\tMYVEC\tMariaDB\tTiDB\n")

	dbs := []*db.MySQLDB{mysqlDB, myvecDB, mariaDB, tiDB}
	for _, col := range []string{"VECTOR", "VECTOR()", "VECTOR(-1)", "VECTOR(0)", "VECTOR(1)", "VECTOR(5)"} {
		fmt.Printf(col)
		for _, cdb := range dbs {
			if cdb.SupportColumn(col) {
				fmt.Printf("\tYES")
			} else {
				fmt.Printf("\tNO")
			}
		}
		fmt.Printf("\n")
	}

	for _, cdb := range dbs {
		_, err = cdb.Exec("DROP TABLE IF EXISTS vt")
		if err != nil {
			log.Fatalf("Failed to drop table: %v", err)
		}
		_, err = cdb.Exec("CREATE TABLE vt (id bigint not null auto_increment primary key, v vector(3), note varchar(255))")
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}

	testVector := []float32{1.1, 2.2, 3.3}
	stringVector := []string{"1.1", "2.2", "3.3", "4.4"}
	testVectorBytes := unsafe.Slice((*byte)(unsafe.Pointer(&testVector[0])), len(testVector)*4)

	for _, cdb := range dbs {
		_, err = cdb.Exec("INSERT INTO vt (v, note) VALUES (?,?)", testVectorBytes, "[3]float32")
		if err != nil {
			fmt.Printf("%s:Binary (%v) not supported: %v\n", cdb.Name, testVectorBytes, err)
		}
	}

	toVecMap := make(map[string]string)
	for _, toVec := range []string{"STRING_TO_VECTOR", "TO_VECTOR", "VEC_FromText", "MYVECTOR_CONSTRUCT", "VEC_FROM_TEXT"} {
		for _, cdb := range dbs {
			_, err = cdb.Exec(fmt.Sprintf("INSERT INTO vt (v, note) VALUES (%s('[1.1,2.2,3.3]'), '%s')", toVec, toVec))
			if err != nil {
				fmt.Printf("%s:%s not supported: %v\n", cdb.Name, toVec, err)
			} else {
				fmt.Printf("%s:%s supported\n", cdb.Name, toVec)
				if _, ok := toVecMap[toVec]; !ok {
					toVecMap[cdb.Name] = toVec
				}
			}
		}
	}
	for _, cdb := range dbs {
		_, err = cdb.Exec(fmt.Sprintf("INSERT INTO vt (v, note) VALUES ('[1.1,2.2,3.3]', '%s')", "Implicit"))
		if err != nil {
			fmt.Printf("%s:%s not supported: %v\n", cdb.Name, "Implicit", err)
		} else {
			fmt.Printf("%s:%s supported\n", cdb.Name, "Implicit")
		}
	}

	for _, cdb := range dbs {
		_, err = cdb.Exec(fmt.Sprintf("INSERT INTO vt (v, note) VALUES (x'CDCC8C3FCDCC0C4033335340', '%s')", "Hex"))
		if err != nil {
			fmt.Printf("%s:%s not supported: %v\n", cdb.Name, "Hex", err)
		} else {
			fmt.Printf("%s:%s supported\n", cdb.Name, "Hex")
		}
	}

	for _, cdb := range dbs {
		_, err = cdb.Exec(fmt.Sprintf("insert into vt (v, note) values (%s('[%s]'), '%s')", toVecMap[cdb.Name], strings.Join(stringVector[:2], ","), "short"))
		if err != nil {
			fmt.Printf("%s not supporting less dim: %v\n", cdb.Name, err)
		} else {
			fmt.Printf("%s is supporting less dim!\n", cdb.Name)
		}
		_, err = cdb.Exec(fmt.Sprintf("insert into vt (v, note) values (%s('[%s]'), '%s')", toVecMap[cdb.Name], strings.Join(stringVector, ","), "long"))
		if err != nil {
			fmt.Printf("%s not supporting more dim: %v\n", cdb.Name, err)
		} else {
			fmt.Printf("%s is supporting more dim!\n", cdb.Name)
		}
	}

	for _, cdb := range dbs {
		rs, err := cdb.Query("SELECT id, v, note FROM vt")
		if err != nil {
			log.Fatalf("%s: Failed to select text vector by function (%s) from table: %v\n", cdb.Name, "Implicit", err)
		}
		for rs.Next() {
			var id int64
			var note string
			var vec []byte
			err = rs.Scan(&id, &vec, &note)
			if err != nil {
				log.Fatalf("%s: Failed to fetch text vector from table: %v\n", cdb.Name, err)
			}
			var str string
			if vec[0] == 91 {
				str = string(vec)
			}
			fmt.Printf("%s returned %v\t%v\t%v (%s)\n", cdb.Name, id, vec, note, str)
		}
	}

	fromVecMap := make(map[string]string)
	for _, fromVec := range []string{"VECTOR_TO_STRING", "FROM_VECTOR", "VEC_ToText", "MYVECTOR_DISPLAY", "VEC_AS_TEXT", "HEX", "VECTOR_DIM", "VEC_DIMS", "VEC_L2_NORM"} {
		for _, cdb := range dbs {
			rs, err := cdb.Query(fmt.Sprintf("SELECT id, %s(v) FROM vt LIMIT 1", fromVec))
			if err != nil {
				fmt.Printf("%s:%s not supported: %v\n", cdb.Name, fromVec, err)
				continue
			}
			if _, ok := fromVecMap[fromVec]; !ok {
				fromVecMap[cdb.Name] = fromVec
			}
			for rs.Next() {
				var id int64
				var vec string
				err = rs.Scan(&id, &vec)
				if err != nil {
					log.Fatalf("%s: Failed to fetch text vector from table: %v\n", cdb.Name, err)
				}
				fmt.Printf("%s:%s returned %v\t%v\n", cdb.Name, fromVec, id, vec)
			}
		}
	}

	distTypes := []string{"COSINE", "EUCLIDEAN", "L2", "DOT", "IP"}
	// "DISTANCE is documented for MySQL, but is not supported in Community or Enterprise version, only Heatwave.
	for _, distFn := range []string{ /*"DISTANCE", */ "MYVECTOR_DISTANCE"} {
		for _, distType := range distTypes {
			for _, cdb := range dbs {
				q := fmt.Sprintf("SELECT id, %s(v, %s('[1.0,1.0,1.0]'), '%s') as 'distance' FROM vt ORDER BY id ASC LIMIT 1", distFn, toVecMap[cdb.Name], distType)
				rs, err := cdb.Query(q)
				if err != nil {
					fmt.Printf("%s:%s+%s not supported: %v\n", cdb.Name, distFn, distType, err)
					continue
				}
				for rs.Next() {
					var id int64
					var d float64
					err = rs.Scan(&id, &d)
					if err != nil {
						fmt.Printf("%s:%s+%s: Failed to fetch text vector from table: %v\n", cdb.Name, distFn, distType, err)
						rs.Close()
						break
					}
					fmt.Printf("%s:%s+%s returned %v\t%v\n", cdb.Name, distFn, distType, id, d)
				}
			}
		}
	}
	for _, distFn := range []string{"VEC_NEGATIVE_INNER_PRODUCT", "VEC_L2_DISTANCE", "VEC_COSINE_DISTANCE", "VEC_DISTANCE_EUCLIDEAN", "VEC_DISTANCE_COSINE", "VEC_DISTANCE"} {
		for _, cdb := range dbs {
			q := fmt.Sprintf("SELECT id, %s(v, %s('[1.0,1.0,1.0]')) as 'distance' FROM vt ORDER BY id ASC LIMIT 1", distFn, toVecMap[cdb.Name])
			rs, err := cdb.Query(q)
			if err != nil {
				fmt.Printf("%s:%s not supported: %v\n", cdb.Name, distFn, err)
				continue
			}
			for rs.Next() {
				var id int64
				var d float64
				err = rs.Scan(&id, &d)
				if err != nil {
					fmt.Printf("%s:%s: Failed to fetch text vector from table: %v\n", cdb.Name, distFn, err)
					rs.Close()
					break
				}
				fmt.Printf("%s:%s returned %v\t%v\n", cdb.Name, distFn, id, d)
			}
		}
	}
	// TODO: Create index and test:
	// ORDER BY vec_dist_fn(v, vec) LIMIT 10;
	// also when a matching row is inserted in the current non-committed transaction
	// also when it is just committed by other transaction.
}
