package database 

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"tickets/entities/models"
	"tickets/entities/events"
	"tickets/events/handler"
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

func (b BookingsRepository) AddBooking(ctx context.Context, booking models.Booking) error {
	return updateInTx(
		ctx,
		b.db,
		sql.LevelSerializable,
		func(ctx context.Context, tx *sqlx.Tx) error {
			availableSeats := 0
			err := tx.GetContext(ctx, &availableSeats, `
				SELECT
					number_of_tickets AS available_seats
				FROM
					shows
				WHERE
					show_id = $1
			`, booking.ShowID)
			if err != nil {
				return fmt.Errorf("could not get available seats: %w", err)
			}

			alreadyBookedSeats := 0
			err = tx.GetContext(ctx, &alreadyBookedSeats, `
				SELECT
					COALESCE(SUM(number_of_tickets), 0) AS already_booked_seats
				FROM
					bookings
				WHERE
					show_id = $1
			`, booking.ShowID)
			if err != nil {
				return fmt.Errorf("could not get already booked seats: %w", err)
			}

			remainingSeats := availableSeats - alreadyBookedSeats
			if remainingSeats < booking.NumberOfTickets {
				return echo.NewHTTPError(
					http.StatusBadRequest,
					fmt.Sprintf("not enough seats available: %d requested but only %d remaining",
						booking.NumberOfTickets, remainingSeats),
				)
			}

			_, err = tx.NamedExecContext(ctx, `
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

			eventBus := handler.NewEventBus(outboxPublisher)

			bookingMadeEvent := events.NewBookingMade(
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
