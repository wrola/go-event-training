package database 

import (
	"context"
	"github.com/jmoiron/sqlx" 
	"tickets/entities"
)

type TicketRepository struct { 
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) TicketRepository { 

	return TicketRepository{db: db}
}

func (t TicketRepository) Add (ctx context.Context, ticket entities.Ticket) (error) {

	_, err := t.db.NamedExecContext(
	ctx,
		`
		INSERT INTO 
    		tickets (ticket_id, price_amount, price_currency, customer_email) 
		VALUES 
		    (:ticket_id, :price.amount, :price.currency, :customer_email)
		ON CONFLICT DO NOTHING;	
		`,
		ticket,
	)
	if err != nil {
		return err
	}

	return nil
}

func (t TicketRepository) Remove (ctx context.Context, ticket entities.Ticket) (error) {
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

func (t TicketRepository) GetAll (ctx context.Context) ([]entities.Ticket, error) {
	var returnTickets []entities.Ticket
	err := t.db.SelectContext(
		ctx,
		&returnTickets, 
		`
			SELECT 
				ticket_id, 
				price_amount as "price.amount",
				price_currency as "price.currency",
				customer_email
			FROM 
				tickets
		`,
	)

	if err != nil {
		return nil, err
	}

	return returnTickets, nil
}