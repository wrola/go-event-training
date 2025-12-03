package database 

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"tickets/entities"
)

type ShowsRepository struct {
	db *sqlx.DB
}

func NewShowsRepository(db *sqlx.DB) ShowsRepository {
	if db == nil {
		panic("db is nil")
	}

	return ShowsRepository{db: db}
}

func (s ShowsRepository) AddShow(ctx context.Context, show entities.Show) error {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO 
		    shows (show_id, dead_nation_id, number_of_tickets, start_time, title, venue) 
		VALUES (:show_id, :dead_nation_id, :number_of_tickets, :start_time, :title, :venue)
		`, show)
	if err != nil {
		return fmt.Errorf("could not add show: %w", err)
	}

	return nil
}

func (s ShowsRepository) ShowByID(ctx context.Context, showID string) (entities.Show, error) {
	var show entities.Show
	err := s.db.GetContext(ctx, &show, `SELECT * FROM shows WHERE show_id = $1`, showID)
	if err != nil {
		return entities.Show{}, err
	}

	return show, nil
}
