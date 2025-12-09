package database

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"tickets/entities/models"
)

type TicketRepository struct { 
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) TicketRepository { 

	return TicketRepository{db: db}
}

func (t TicketRepository) Add (ctx context.Context, ticket models.Ticket) (error) {
	confirmedAt := sql.NullString{String: ticket.ConfirmedAt, Valid: ticket.ConfirmedAt != ""}
	refundedAt := sql.NullString{String: ticket.RefundedAt, Valid: ticket.RefundedAt != ""}

	_, err := t.db.ExecContext(
		ctx,
		`INSERT INTO tickets (ticket_id, price_amount, price_currency, customer_email, confirmed_at, refunded_at)
		 VALUES ($1, $2, $3, $4, $5::TIMESTAMPTZ, $6::TIMESTAMPTZ)
		 ON CONFLICT DO NOTHING`,
		ticket.TicketID,
		ticket.Price.Amount,
		ticket.Price.Currency,
		ticket.CustomerEmail,
		confirmedAt,
		refundedAt,
	)
	return err
}

func (t TicketRepository) Remove (ctx context.Context, ticket models.Ticket) (error) {
	_, err := t.db.NamedExecContext(
		ctx, 
		`DELETE FROM 
			tickets as t
		WHERE
			t.ticket_id = :ticket_id`,
		ticket,
	)
	if err != nil {
		return err
	}

	return nil
}

func (t TicketRepository) UpdateRefundedAt(ctx context.Context, ticketID string, refundedAt string) error {
	refundedAtNull := sql.NullString{String: refundedAt, Valid: refundedAt != ""}

	_, err := t.db.ExecContext(
		ctx,
		`UPDATE tickets
		 SET refunded_at = $1::TIMESTAMPTZ
		 WHERE ticket_id = $2`,
		refundedAtNull,
		ticketID,
	)
	return err
}

func (t TicketRepository) GetAll (ctx context.Context) ([]models.Ticket, error) {
	var returnTickets []models.Ticket
	err := t.db.SelectContext(
		ctx,
		&returnTickets,
		`
			SELECT
				ticket_id,
				price_amount as "price.amount",
				price_currency as "price.currency",
				customer_email,
				COALESCE(confirmed_at::TEXT, '') as confirmed_at,
				COALESCE(refunded_at::TEXT, '') as refunded_at
			FROM
				tickets
		`,
	)

	if err != nil {
		return nil, err
	}

	return returnTickets, nil
}