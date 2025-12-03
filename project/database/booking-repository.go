package database 

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"tickets/entities"
)

type BookingsRepository struct {
	db *sqlx.DB
}

func NewBookingsRepository(db *sqlx.DB) BookingsRepository {
	if db == nil {
		panic("db is nil")
	}

	return BookingsRepository{db: db}
}

func (s BookingsRepository) AddBooking(ctx context.Context, booking entities.Booking) error {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO
		    bookings (booking_id, show_id, number_of_tickets, customer_email)
		VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
		`, booking)
	if err != nil {
		return fmt.Errorf("could not add booking: %w", err)
	}

	return nil
}