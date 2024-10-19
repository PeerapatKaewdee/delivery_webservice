package config

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() {
	dsn := "web66_65011212243:65011212243@csmsu@tcp(202.28.34.197:3306)/web66_65011212243"
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Database is unreachable:", err)
	}

	log.Println("Database connection successful")
}
