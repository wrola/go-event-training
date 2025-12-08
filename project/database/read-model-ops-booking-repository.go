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

type sqlExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type OpsBookingReadModelRepository struct {
	db *sqlx.DB
}

func NewOpsBookingReadModelRespository(db *sqlx.DB) OpsBookingReadModelRepository {
	if db == nil {
		panic("db is nil")
	}
	return OpsBookingReadModelRepository{db: db}
}

func (r OpsBookingReadModelRepository) CreateBooking(
	ctx context.Context,
	bookingID string,
	bookedAt time.Time,
) error {
	parsedBookingID, err := uuid.Parse(bookingID)
	if err != nil {
		return fmt.Errorf("invalid booking_id: %w", err)
	}

	return updateInTx(ctx, r.db, sql.LevelRepeatableRead, func(ctx context.Context, tx *sqlx.Tx) error {
		return r.createReadModel(ctx, tx, parsedBookingID, bookedAt)
	})
}

func (r OpsBookingReadModelRepository) createReadModel(
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

func (r OpsBookingReadModelRepository) UpdateReadModelByBookingID(
	ctx context.Context,
	bookingID string,
	updateFn func(*models.OpsBooking) error,
) error {
	return updateInTx(ctx, r.db, sql.LevelReadCommitted, func(ctx context.Context, tx *sqlx.Tx) error {
		opsBooking, err := r.findReadModelByBookingID(ctx, tx, bookingID, true)
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

func (r OpsBookingReadModelRepository) UpdateReadModelByTicketID(
	ctx context.Context,
	ticketID string,
	updateFn func(*models.OpsBooking) error,
) error {
	return updateInTx(ctx, r.db, sql.LevelReadCommitted, func(ctx context.Context, tx *sqlx.Tx) error {
		opsBooking, err := r.findReadModelByTicketID(ctx, tx, ticketID, true)
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

func (r OpsBookingReadModelRepository) FindReadModelByBookingID(
	ctx context.Context,
	bookingID string,
) (models.OpsBooking, error) {
	return r.findReadModelByBookingID(ctx, r.db, bookingID, false)
}

func (r OpsBookingReadModelRepository) findReadModelByBookingID(
	ctx context.Context,
	exec sqlExecutor,
	bookingID string,
	forUpdate bool,
) (models.OpsBooking, error) {
	var payloadBytes []byte

	query := `
		SELECT payload
		FROM read_model_ops_bookings
		WHERE booking_id = $1
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	err := exec.QueryRowContext(ctx, query, bookingID).Scan(&payloadBytes)

	if err == sql.ErrNoRows {
		return models.OpsBooking{}, fmt.Errorf("ops booking not found for booking_id %s: %w", bookingID, err)
	}
	if err != nil {
		return models.OpsBooking{}, fmt.Errorf("failed to query ops booking: %w", err)
	}

	return r.unmarshalReadModelFromDB(payloadBytes)
}

func (r OpsBookingReadModelRepository) FindReadModelByTicketID(
	ctx context.Context,
	ticketID string,
) (models.OpsBooking, error) {
	return r.findReadModelByTicketID(ctx, r.db, ticketID, false)
}

func (r OpsBookingReadModelRepository) findReadModelByTicketID(
	ctx context.Context,
	exec sqlExecutor,
	ticketID string,
	forUpdate bool,
) (models.OpsBooking, error) {
	var payloadBytes []byte

	query := `
		SELECT payload
		FROM read_model_ops_bookings
		WHERE payload -> 'tickets' ? $1
	`
	if forUpdate {
		query += " FOR UPDATE"
	}

	err := exec.QueryRowContext(ctx, query, ticketID).Scan(&payloadBytes)

	if err == sql.ErrNoRows {
		return models.OpsBooking{}, fmt.Errorf("ops booking not found for ticket_id %s (ticket may not exist yet): %w", ticketID, err)
	}
	if err != nil {
		return models.OpsBooking{}, fmt.Errorf("failed to query ops booking by ticket: %w", err)
	}

	return r.unmarshalReadModelFromDB(payloadBytes)
}

func (r OpsBookingReadModelRepository) GetAllReadModels(
	ctx context.Context,
	receiptIssueDate *time.Time,
) ([]models.OpsBooking, error) {

	var payloadBytes [][]byte

	query := `
		SELECT payload
		FROM read_model_ops_bookings
	`

	var args []interface{}

	if receiptIssueDate != nil {
		query += `
		WHERE booking_id IN (
			SELECT booking_id FROM (
				SELECT booking_id,
					DATE(jsonb_path_query(payload, '$.tickets.*.receipt_issued_at')::text) as receipt_issued_at
				FROM read_model_ops_bookings
			) bookings_within_date
			WHERE receipt_issued_at = $1
		)
		`
		args = append(args, receiptIssueDate.Format("2006-01-02"))
	}

	err := r.db.SelectContext(
		ctx,
		&payloadBytes,
		query,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ops bookings: %w", err)
	}

	return r.unmarshalReadModelListFromDB(payloadBytes)
}

func (r OpsBookingReadModelRepository) UpdateReadModel(
	ctx context.Context,
	rm models.OpsBooking,
) error {
	return r.updateReadModel(ctx, r.db, rm)
}

func (r OpsBookingReadModelRepository) updateReadModel(
	ctx context.Context,
	exec sqlExecutor,
	rm models.OpsBooking,
) error {
	rm.LastUpdate = time.Now()

	payload, err := json.Marshal(rm)
	if err != nil {
		return fmt.Errorf("failed to marshal ops booking: %w", err)
	}

	_, err = exec.ExecContext(ctx, `
		UPDATE read_model_ops_bookings
		SET payload = $2
		WHERE booking_id = $1
	`, rm.BookingID, payload)

	if err != nil {
		return fmt.Errorf("failed to update ops booking: %w", err)
	}

	return nil
}

func (r OpsBookingReadModelRepository) unmarshalReadModelFromDB(payloadBytes []byte) (models.OpsBooking, error) {
	var opsBooking models.OpsBooking
	if err := json.Unmarshal(payloadBytes, &opsBooking); err != nil {
		return models.OpsBooking{}, fmt.Errorf("failed to unmarshal ops booking: %w", err)
	}

	if opsBooking.Tickets == nil {
		opsBooking.Tickets = make(map[string]models.OpsTicket)
	}

	return opsBooking, nil
}

func (r OpsBookingReadModelRepository) unmarshalReadModelListFromDB(payloadBytesList [][]byte) ([]models.OpsBooking, error) {
	var opsBookings []models.OpsBooking
	for _, payloadBytes := range payloadBytesList {
		opsBooking, err := r.unmarshalReadModelFromDB(payloadBytes)
		if err != nil {
			return nil, err
		}
		opsBookings = append(opsBookings, opsBooking)
	}
	return opsBookings, nil
}
