package main

import (
	"auth/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "8080"

var counts int64

type App struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication server...")

	//connect to database
	conn := connectToPostgres()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}

	defer conn.Close()

	//set up the application
	app := &App{
		DB:     conn,
		Models: data.New(conn),
	}

	log.Printf("Strating authentication server on port %s\n", webPort)

	//setup the server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Authentication server started successfully!")

}

func openPostgresConn(dsn string) (*sql.DB, error) {
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

func connectToPostgres() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openPostgresConn(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}
