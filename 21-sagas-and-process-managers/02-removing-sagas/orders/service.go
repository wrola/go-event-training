package orders

import (
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func Initialize(e *echo.Echo) {
	db, err := sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(err)
	}

	initializeDatabaseSchema(db)
	mountHttpHandlers(e, db)
}
