package test_ticktet_repository

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tickets/database"
	"tickets/entities/models"
)

var db *sqlx.DB
var getDbOnce sync.Once

func getDb() *sqlx.DB {
	getDbOnce.Do(func() {
		var err error
		db, err = sqlx.Open("postgres", os.Getenv("POSTGRES_URL"))
		if err != nil {
			panic(err)
		}
		
		database.InitializeSchema(db)
	})
	return db
}

func createTestTicket(ticketID string) models.Ticket {
	return models.Ticket{
		TicketID:      ticketID,
		CustomerEmail: "test-" + ticketID + "@example.com",
		Price: models.Money{
			Amount:   "99.99",
			Currency: "USD",
		},
	}
}

func findTicketByID(tickets []models.Ticket, ticketID string) (*models.Ticket, bool) {
	for _, ticket := range tickets {
		if ticket.TicketID == ticketID {
			return &ticket, true
		}
	}
	return nil, false
}


func countTicketsWithID(tickets []models.Ticket, ticketID string) int {
	count := 0
	for _, ticket := range tickets {
		if ticket.TicketID == ticketID {
			count++
		}
	}
	return count
}


func TestTicketRepository_Add(t *testing.T) {
	db := getDb()
	repo := database.NewTicketRepository(db)
	ctx := context.Background()

	ticketID := uuid.New().String()
	ticket := createTestTicket(ticketID)

	err := repo.Add(ctx, ticket)
	require.NoError(t, err, "Add should succeed")

	tickets, err := repo.GetAll(ctx)
	require.NoError(t, err, "GetAll should succeed")

	foundTicket, found := findTicketByID(tickets, ticketID)
	require.True(t, found, "ticket should be found after Add")

	assert.Equal(t, ticket.TicketID, foundTicket.TicketID)
	assert.Equal(t, ticket.CustomerEmail, foundTicket.CustomerEmail)
	assert.Equal(t, ticket.Price.Amount, foundTicket.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, foundTicket.Price.Currency)
}


func TestTicketRepository_AddTwice_Idempotency(t *testing.T) {
	db := getDb()
	repo := database.NewTicketRepository(db)
	ctx := context.Background()

	ticketID := uuid.New().String()
	ticket := createTestTicket(ticketID)

	err := repo.Add(ctx, ticket)
	require.NoError(t, err, "first Add should succeed")

	tickets, err := repo.GetAll(ctx)
	require.NoError(t, err)
	count := countTicketsWithID(tickets, ticketID)
	require.Equal(t, 1, count, "should have exactly 1 ticket after first add")

	err = repo.Add(ctx, ticket)
	require.NoError(t, err, "second Add should succeed (ON CONFLICT DO NOTHING)")

	tickets, err = repo.GetAll(ctx)
	require.NoError(t, err)
	count = countTicketsWithID(tickets, ticketID)
	assert.Equal(t, 1, count, "should STILL have exactly 1 ticket after second add - idempotency!")

	foundTicket, found := findTicketByID(tickets, ticketID)
	require.True(t, found)
	assert.Equal(t, ticket.CustomerEmail, foundTicket.CustomerEmail)
	assert.Equal(t, ticket.Price.Amount, foundTicket.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, foundTicket.Price.Currency)
}

func TestTicketRepository_Remove(t *testing.T) {
	db := getDb()
	repo := database.NewTicketRepository(db)
	ctx := context.Background()

	ticketID := uuid.New().String()
	ticket := createTestTicket(ticketID)

	err := repo.Add(ctx, ticket)
	require.NoError(t, err)

	tickets, err := repo.GetAll(ctx)
	require.NoError(t, err)
	_, found := findTicketByID(tickets, ticketID)
	require.True(t, found, "ticket should exist before removal")

	err = repo.Remove(ctx, ticket)
	require.NoError(t, err, "Remove should succeed")

	tickets, err = repo.GetAll(ctx)
	require.NoError(t, err)
	_, found = findTicketByID(tickets, ticketID)
	assert.False(t, found, "ticket should not exist after removal")
}

func TestTicketRepository_GetAll(t *testing.T) {
	db := getDb()
	repo := database.NewTicketRepository(db)
	ctx := context.Background()

	ticketIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		ticketID := uuid.New().String()
		ticketIDs[i] = ticketID
		ticket := createTestTicket(ticketID)
		err := repo.Add(ctx, ticket)
		require.NoError(t, err)
	}

	tickets, err := repo.GetAll(ctx)
	require.NoError(t, err)

	for _, ticketID := range ticketIDs {
		foundTicket, found := findTicketByID(tickets, ticketID)
		assert.True(t, found, "ticket %s should be found", ticketID)
		if found {
			expectedTicket := createTestTicket(ticketID)
			assert.Equal(t, expectedTicket.CustomerEmail, foundTicket.CustomerEmail)
			assert.Equal(t, expectedTicket.Price.Amount, foundTicket.Price.Amount)
			assert.Equal(t, expectedTicket.Price.Currency, foundTicket.Price.Currency)
		}
	}
}

func TestTicketRepository_RemoveNonExistent(t *testing.T) {
	db := getDb()
	repo := database.NewTicketRepository(db)
	ctx := context.Background()

	ticketID := uuid.New().String()
	ticket := createTestTicket(ticketID)

	err := repo.Remove(ctx, ticket)
	assert.NoError(t, err, "Remove of non-existent ticket should not error")
}

func TestTicketRepository_MultipleDuplicates(t *testing.T) {
	db := getDb()
	repo := database.NewTicketRepository(db)
	ctx := context.Background()

	ticketA := createTestTicket(uuid.New().String())
	ticketB := createTestTicket(uuid.New().String())
	ticketC := createTestTicket(uuid.New().String())
	ticketD := createTestTicket(uuid.New().String())

	require.NoError(t, repo.Add(ctx, ticketA))
	require.NoError(t, repo.Add(ctx, ticketB))
	require.NoError(t, repo.Add(ctx, ticketC))

	require.NoError(t, repo.Add(ctx, ticketA))

	require.NoError(t, repo.Add(ctx, ticketB))

	require.NoError(t, repo.Add(ctx, ticketD))

	tickets, err := repo.GetAll(ctx)
	require.NoError(t, err)

	countA := countTicketsWithID(tickets, ticketA.TicketID)
	countB := countTicketsWithID(tickets, ticketB.TicketID)
	countC := countTicketsWithID(tickets, ticketC.TicketID)
	countD := countTicketsWithID(tickets, ticketD.TicketID)

	assert.Equal(t, 1, countA, "ticket A should appear exactly once")
	assert.Equal(t, 1, countB, "ticket B should appear exactly once")
	assert.Equal(t, 1, countC, "ticket C should appear exactly once")
	assert.Equal(t, 1, countD, "ticket D should appear exactly once")
}
