package entities

import (
	"time"

	"github.com/google/uuid"
)

type Show struct {
    ShowID          string    `json:"show_id" db:"show_id"`
    DeadNationID    string    `json:"dead_nation_id" db:"dead_nation_id"`
    NumberOfTickets int       `json:"number_of_tickets" db:"number_of_tickets"`
    StartTime       time.Time `json:"start_time" db:"start_time"`
    Title           string    `json:"title" db:"title"`
    Venue           string    `json:"venue" db:"venue"`
}

func NewShow(showID, deadNationID string, numberOfTickets int, startTime time.Time, title, venue string) Show {
	if showID == "" {
		showID = uuid.NewString()
	}

	return Show{
		ShowID:          showID,
		DeadNationID:    deadNationID,
		NumberOfTickets: numberOfTickets,
		StartTime:       startTime,
		Title:           title,
		Venue:           venue,
	}
}
