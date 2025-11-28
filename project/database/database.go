package database

import (
	"os"
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx" 
)

func NewDatabaseConnection () (db *sqlx.DB) {
	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}

	return db
}

func InitializeSchema(db *sqlx.DB) {
	_, err := db.Exec(schema)

	if err != nil {
		panic(err)
	}
}

var schema = `CREATE TABLE IF NOT EXISTS tickets (
	ticket_id UUID PRIMARY KEY,
	price_amount NUMERIC(10, 2) NOT NULL,
	price_currency CHAR(3) NOT NULL,
	customer_email VARCHAR(255) NOT NULL
);`