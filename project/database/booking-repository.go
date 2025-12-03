package database 

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/jmoiron/sqlx"

	"tickets/entities"
	"tickets/message/event"
)

func updateInTx(
	ctx context.Context,
	db *sqlx.DB,
	isolation sql.IsolationLevel,
	fn func(ctx context.Context, tx *sqlx.Tx) error,
) (err error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: isolation})
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
			return
		}

		err = tx.Commit()
	}()

	return fn(ctx, tx)
}

type BookingsRepository struct {
	db     *sqlx.DB
	logger watermill.LoggerAdapter
}

func NewBookingsRepository(db *sqlx.DB, logger watermill.LoggerAdapter) BookingsRepository {
	if db == nil {
		panic("db is nil")
	}

	return BookingsRepository{
		db:     db,
		logger: logger,
	}
}

func (b BookingsRepository) AddBooking(ctx context.Context, booking entities.Booking) error {
	return updateInTx(
		ctx,
		b.db,
		sql.LevelRepeatableRead,
		func(ctx context.Context, tx *sqlx.Tx) error {
			_, err := tx.NamedExecContext(ctx, `
				INSERT INTO
					bookings (booking_id, show_id, number_of_tickets, customer_email)
				VALUES (:booking_id, :show_id, :number_of_tickets, :customer_email)
			`, booking)
			if err != nil {
				return fmt.Errorf("could not add booking: %w", err)
			}

			outboxPublisher, err := NewOutboxPublisher(tx, b.logger)
			if err != nil {
				return fmt.Errorf("could not create outbox publisher: %w", err)
			}

			eventBus := event.NewEventBus(outboxPublisher)

			bookingMadeEvent := entities.NewBookingMade(
				booking.BookingID,
				booking.CustomerEmail,
				booking.ShowID,
				booking.NumberOfTickets,
				booking.BookingID, 
			)

			err = eventBus.Publish(ctx, bookingMadeEvent)
			if err != nil {
				return fmt.Errorf("could not publish BookingMade event: %w", err)
			}

			return nil
		},
	)
}