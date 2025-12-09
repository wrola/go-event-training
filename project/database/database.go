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
	customer_email VARCHAR(255) NOT NULL,
	confirmed_at TIMESTAMPTZ,
	refunded_at TIMESTAMPTZ,
	deleted_at TIMESTAMPTZ
);
	CREATE TABLE IF NOT EXISTS shows (
	  show_id UUID PRIMARY KEY,
	  dead_nation_id UUID NOT NULL,
	  number_of_tickets INT NOT NULL,
	  start_time TIMESTAMP NOT NULL,
	  title VARCHAR(255) NOT NULL,
	  venue VARCHAR(255) NOT NULL,

	  UNIQUE (dead_nation_id)
);

CREATE TABLE IF NOT EXISTS bookings (
    	booking_id UUID PRIMARY KEY,
    	show_id UUID NOT NULL,
    	number_of_tickets INT NOT NULL,
    	customer_email VARCHAR(255) NOT NULL,
    	FOREIGN KEY (show_id) REFERENCES shows(show_id)
);

CREATE TABLE IF NOT EXISTS read_model_ops_bookings (
    booking_id UUID PRIMARY KEY,
    payload JSONB NOT NULL
);
`