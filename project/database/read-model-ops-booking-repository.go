package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"tickets/entities/models"
)

type OpsBookingReadModel struct {
	db *sqlx.DB
}

func NewOpsBookingReadModel(db *sqlx.DB) OpsBookingReadModel {
	if db == nil {
		panic("db is nil")
	}
	return OpsBookingReadModel{db: db}
}

func (r OpsBookingReadModel) CreateBooking(
	ctx context.Context,
	bookingID uuid.UUID,
	bookedAt time.Time,
) error {
	return updateInTx(ctx, r.db, sql.LevelRepeatableRead, func(ctx context.Context, tx *sqlx.Tx) error {
		return r.createReadModel(ctx, tx, bookingID, bookedAt)
	})
}

func (r OpsBookingReadModel) createReadModel(
	ctx context.Context,
	tx *sqlx.Tx,
	bookingID uuid.UUID,
	bookedAt time.Time,
) error {
	opsBooking := models.OpsBooking{
		BookingID:  bookingID,
		BookedAt:   bookedAt,
		Tickets:    make(map[string]models.OpsTicket),
		LastUpdate: time.Now(),
	}

	payload, err := json.Marshal(opsBooking)
	if err != nil {
		return fmt.Errorf("failed to marshal ops booking: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO read_model_ops_bookings (booking_id, payload)
		VALUES ($1, $2)
		ON CONFLICT (booking_id) DO NOTHING
	`, bookingID, payload)

	if err != nil {
		return fmt.Errorf("failed to insert ops booking: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) UpdateReadModelByBookingID(
	ctx context.Context,
	bookingID uuid.UUID,
	updateFn func(*models.OpsBooking) error,
) error {
	return updateInTx(ctx, r.db, sql.LevelRepeatableRead, func(ctx context.Context, tx *sqlx.Tx) error {
		opsBooking, err := r.findReadModelByBookingID(ctx, tx, bookingID)
		if err != nil {
			return err
		}

		if err := updateFn(&opsBooking); err != nil {
			return err
		}

		opsBooking.LastUpdate = time.Now()

		return r.updateReadModel(ctx, tx, opsBooking)
	})
}

func (r OpsBookingReadModel) UpdateReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	updateFn func(*models.OpsBooking) error,
) error {
	return updateInTx(ctx, r.db, sql.LevelRepeatableRead, func(ctx context.Context, tx *sqlx.Tx) error {
		opsBooking, err := r.findReadModelByTicketID(ctx, tx, ticketID)
		if err != nil {
			return err
		}

		if err := updateFn(&opsBooking); err != nil {
			return err
		}

		opsBooking.LastUpdate = time.Now()

		return r.updateReadModel(ctx, tx, opsBooking)
	})
}

func (r OpsBookingReadModel) findReadModelByBookingID(
	ctx context.Context,
	tx *sqlx.Tx,
	bookingID uuid.UUID,
) (models.OpsBooking, error) {
	var payloadBytes []byte
	err := tx.QueryRowContext(ctx, `
		SELECT payload
		FROM read_model_ops_bookings
		WHERE booking_id = $1
	`, bookingID).Scan(&payloadBytes)

	if err == sql.ErrNoRows {
		return models.OpsBooking{}, fmt.Errorf("ops booking not found for booking_id %s: %w", bookingID, err)
	}
	if err != nil {
		return models.OpsBooking{}, fmt.Errorf("failed to query ops booking: %w", err)
	}

	return r.unmarshalReadModelFromDB(payloadBytes)
}

func (r OpsBookingReadModel) findReadModelByTicketID(
	ctx context.Context,
	tx *sqlx.Tx,
	ticketID string,
) (models.OpsBooking, error) {
	var payloadBytes []byte

	err := tx.QueryRowContext(ctx, `
		SELECT payload
		FROM read_model_ops_bookings
		WHERE payload -> 'tickets' ? $1
	`, ticketID).Scan(&payloadBytes)

	if err == sql.ErrNoRows {
		return models.OpsBooking{}, fmt.Errorf("ops booking not found for ticket_id %s (ticket may not exist yet): %w", ticketID, err)
	}
	if err != nil {
		return models.OpsBooking{}, fmt.Errorf("failed to query ops booking by ticket: %w", err)
	}

	return r.unmarshalReadModelFromDB(payloadBytes)
}

func (r OpsBookingReadModel) updateReadModel(
	ctx context.Context,
	tx *sqlx.Tx,
	rm models.OpsBooking,
) error {
	rm.LastUpdate = time.Now()

	payload, err := json.Marshal(rm)
	if err != nil {
		return fmt.Errorf("failed to marshal ops booking: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE read_model_ops_bookings
		SET payload = $2
		WHERE booking_id = $1
	`, rm.BookingID, payload)

	if err != nil {
		return fmt.Errorf("failed to update ops booking: %w", err)
	}

	return nil
}

func (r OpsBookingReadModel) unmarshalReadModelFromDB(payloadBytes []byte) (models.OpsBooking, error) {
	var opsBooking models.OpsBooking
	if err := json.Unmarshal(payloadBytes, &opsBooking); err != nil {
		return models.OpsBooking{}, fmt.Errorf("failed to unmarshal ops booking: %w", err)
	}

	if opsBooking.Tickets == nil {
		opsBooking.Tickets = make(map[string]models.OpsTicket)
	}

	return opsBooking, nil
}
