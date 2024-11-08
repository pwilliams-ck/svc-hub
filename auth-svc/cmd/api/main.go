package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/cloudkey-io/service-hub/auth-svc/data"
)

const port = "80"

var counts int64

type application struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service...")

	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	// Set up config
	app := application{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Connects to DB, waits for DB connection if it is taking a while.
func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	counts := 0
	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not ready yet...")
			counts++
		} else {
			log.Println("Connected to Postgres...")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for 2 seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}
