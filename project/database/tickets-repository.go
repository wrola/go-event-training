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
		    (:ticket_id, :price.amount, :price.currency, :customer_email)`,
		ticket,
	)
	if err != nil {
		return err
	}

	return nil
}