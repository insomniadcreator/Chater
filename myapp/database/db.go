package database

import (
    "database/sql"
    _ "github.com/lib/pq" // для PostgreSQL
    // _ "github.com/go-sql-driver/mysql" // для MySQL
    "log"
)

var DB *sql.DB

func InitDB(connStr string) {
    var err error
    DB, err = sql.Open("postgres", connStr) // или "mysql" для MySQL
    if err != nil {
        log.Fatal(err)
    }

    err = DB.Ping()
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Connected to the database")
}