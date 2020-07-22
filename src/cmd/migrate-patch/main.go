package main

import (
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"log"
	"strings"
	"time"
)

var dbURL string

const pgSQLAlterStmt string = `ALTER TABLE schema_migrations ADD COLUMN "dirty" boolean NOT NULL DEFAULT false`
const pgSQLCheckColStmt string = `SELECT T1.C1, T2.C2 FROM
(SELECT COUNT(*) AS C1 FROM information_schema.tables WHERE table_name='schema_migrations') T1,
(SELECT COUNT(*) AS C2 FROM information_schema.columns WHERE table_name='schema_migrations' and column_name='dirty') T2`
const pgSQLDelRows string = `DELETE FROM schema_migrations t WHERE t.version < ( SELECT MAX(version) FROM schema_migrations )`

func init() {
	urlUsage := `The URL to the target database (driver://url).  Currently it only supports postgres`
	flag.StringVar(&dbURL, "database", "", urlUsage)
}

func main() {
	flag.Parse()
	log.Printf("Updating database.")
	if !strings.HasPrefix(dbURL, "postgres://") {
		log.Fatalf("Invalid URL: '%s'\n", dbURL)
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to Database, error: %v\n", err)
	}
	defer db.Close()
	c := make(chan struct{}, 1)
	go func() {
		err := db.Ping()
		for ; err != nil; err = db.Ping() {
			log.Println("Failed to Ping DB, sleep for 1 second.")
			time.Sleep(1 * time.Second)
		}
		c <- struct{}{}
	}()
	select {
	case <-c:
	case <-time.After(30 * time.Second):
		log.Fatal("Failed to connect DB after 30 seconds, time out. \n")

	}
	row := db.QueryRow(pgSQLCheckColStmt)
	var tblCount, colCount int
	if err := row.Scan(&tblCount, &colCount); err != nil {
		log.Fatalf("Failed to check schema_migrations table, error: %v \n", err)
	}
	if tblCount == 0 {
		log.Println("schema_migrations table does not exist, skip.")
		return
	}
	if colCount > 0 {
		log.Println("schema_migrations table does not require update, skip.")
		return
	}
	if _, err := db.Exec(pgSQLDelRows); err != nil {
		log.Fatalf("Failed to clean up table, error: %v", err)
	}
	if _, err := db.Exec(pgSQLAlterStmt); err != nil {
		log.Fatalf("Failed to update database, error: %v \n", err)
	}
	log.Println("Done updating database.")
}
