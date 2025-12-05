package tests_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"tickets/adapters"
	"tickets/database"
	"tickets/entities"
	ticketsHttp "tickets/http"
	"tickets/service"
	"github.com/lithammer/shortuuid/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	os.Setenv("PORT", "8080")
	if os.Getenv("REDIS_ADDR") == "" {
		os.Setenv("REDIS_ADDR", "localhost:6379")
	}

	spreadsheetsAPI := &adapters.SpreadsheetsAPIClientStub{}
	receiptsService := &adapters.ReceiptsServiceClientStub{}
	paymentsService := &adapters.PaymentsServiceClientStub{}
	fileAPI := &adapters.FilesAPIClientStub{}
	deadNationAPI := &adapters.DeadNationAPIStub{}
	db := database.NewDatabaseConnection()

	go func() {
		svc, err := service.New(
			spreadsheetsAPI,
			receiptsService,
			paymentsService,
			fileAPI,
			deadNationAPI,
			db,
		)
		if err != nil {
			t.Errorf("service.New() error: %v", err)
			return
		}

		err = svc.RunWithGracefulShutdown("")
		if err != nil {
			t.Errorf("service.Run() error: %v", err)
		}
	}()

	waitForHttpServer(t)

	t.Run("confirmed ticket", func(t *testing.T) {
		ticket := ticketsHttp.TicketStatusRequest{
			TicketID:      "test-ticket-confirmed-123",
			CustomerEmail: "confirmed@example.com",
			Price: entities.Money{
				Amount:   "100.00",
				Currency: "USD",
			},
			Status: entities.TicketStatusConfirmed,
		}

		sendTicketsStatus(t, ticket)

		assertReceiptForTicketIssued(t, receiptsService, ticket)

		assertRowAddedToSpreadsheet(t, spreadsheetsAPI, ticket, "tickets-to-print")
	})

	t.Run("canceled ticket", func(t *testing.T) {
		ticket := ticketsHttp.TicketStatusRequest{
			TicketID:      "test-ticket-canceled-456",
			CustomerEmail: "canceled@example.com",
			Price: entities.Money{
				Amount:   "50.00",
				Currency: "EUR",
			},
			Status: entities.TicketStatusCanceled,
		}

		sendTicketsStatus(t, ticket)

		assertRowAddedToSpreadsheet(t, spreadsheetsAPI, ticket, "tickets-to-refund")
	})

	t.Run("get all tickets", func(t *testing.T) {
		ticket1 := ticketsHttp.TicketStatusRequest{
			TicketID:      "get-all-test-1",
			CustomerEmail: "test1@example.com",
			Price: entities.Money{
				Amount:   "25.50",
				Currency: "USD",
			},
			Status: entities.TicketStatusConfirmed,
		}

		ticket2 := ticketsHttp.TicketStatusRequest{
			TicketID:      "get-all-test-2",
			CustomerEmail: "test2@example.com",
			Price: entities.Money{
				Amount:   "35.75",
				Currency: "EUR",
			},
			Status: entities.TicketStatusConfirmed,
		}

		sendTicketsStatus(t, ticket1)
		sendTicketsStatus(t, ticket2)

		time.Sleep(1 * time.Second)

		resp, err := http.Get("http://localhost:8080/tickets")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var tickets []entities.Ticket
		err = json.NewDecoder(resp.Body).Decode(&tickets)
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(tickets), 2, "should have at least 2 tickets")

		var foundTicket1, foundTicket2 bool
		for _, ticket := range tickets {
			if ticket.TicketID == ticket1.TicketID {
				foundTicket1 = true
				assert.Equal(t, ticket1.CustomerEmail, ticket.CustomerEmail)
				assert.Equal(t, ticket1.Price.Amount, ticket.Price.Amount)
				assert.Equal(t, ticket1.Price.Currency, ticket.Price.Currency)
			}
			if ticket.TicketID == ticket2.TicketID {
				foundTicket2 = true
				assert.Equal(t, ticket2.CustomerEmail, ticket.CustomerEmail)
				assert.Equal(t, ticket2.Price.Amount, ticket.Price.Amount)
				assert.Equal(t, ticket2.Price.Currency, ticket.Price.Currency)
			}
		}

		assert.True(t, foundTicket1, "ticket1 should be in the response")
		assert.True(t, foundTicket2, "ticket2 should be in the response")
	})

	t.Run("booking within capacity", func(t *testing.T) {
		showID := "test-show-" + shortuuid.New()

		createShow(t, showID, 10)

		resp := bookTickets(t, showID, 5, "user1@example.com")
		assertBookingSucceeds(t, resp)

		resp = bookTickets(t, showID, 3, "user2@example.com")
		assertBookingSucceeds(t, resp)
	})

	t.Run("overbooking attempt fails", func(t *testing.T) {
		showID := "test-show-" + shortuuid.New()

		createShow(t, showID, 10)

		resp := bookTickets(t, showID, 8, "user1@example.com")
		assertBookingSucceeds(t, resp)

		resp = bookTickets(t, showID, 5, "user2@example.com")
		assertBookingFails(t, resp, "not enough seats available")
	})

	t.Run("exact capacity booking", func(t *testing.T) {
		showID := "test-show-" + shortuuid.New()

		createShow(t, showID, 10)

		resp := bookTickets(t, showID, 6, "user1@example.com")
		assertBookingSucceeds(t, resp)

		resp = bookTickets(t, showID, 4, "user2@example.com")
		assertBookingSucceeds(t, resp)

		// Try to book 1 more ticket (no capacity left, should fail)
		resp = bookTickets(t, showID, 1, "user3@example.com")
		assertBookingFails(t, resp, "not enough seats available")
	})

	t.Run("first booking for show", func(t *testing.T) {
		showID := "test-show-" + shortuuid.New()

		createShow(t, showID, 100)

		resp := bookTickets(t, showID, 20, "user1@example.com")
		assertBookingSucceeds(t, resp)
	})

	t.Run("zero tickets remaining", func(t *testing.T) {
		showID := "test-show-" + shortuuid.New()

		createShow(t, showID, 5)

		resp := bookTickets(t, showID, 5, "user1@example.com")
		assertBookingSucceeds(t, resp)

		resp = bookTickets(t, showID, 1, "user2@example.com")
		assertBookingFails(t, resp, "not enough seats available")
	})

	t.Run("concurrent bookings prevent overbooking", func(t *testing.T) {
		showID := "test-show-" + shortuuid.New()

		createShow(t, showID, 10)

		var wg sync.WaitGroup
		results := make(chan *http.Response, 2)

		wg.Add(2)

		go func() {
			defer wg.Done()
			resp := bookTickets(t, showID, 6, "concurrent1@example.com")
			results <- resp
		}()

		go func() {
			defer wg.Done()
			resp := bookTickets(t, showID, 6, "concurrent2@example.com")
			results <- resp
		}()

		wg.Wait()
		close(results)

		var responses []*http.Response
		for resp := range results {
			responses = append(responses, resp)
		}

		require.Len(t, responses, 2)

		successCount := 0
		failureCount := 0

		for _, resp := range responses {
			if resp.StatusCode == http.StatusCreated {
				successCount++
			} else if resp.StatusCode == http.StatusBadRequest {
				failureCount++
			}
			resp.Body.Close()
		}

		assert.Equal(t, 1, successCount, "exactly one concurrent booking should succeed")
		assert.Equal(t, 1, failureCount, "exactly one concurrent booking should fail")
	})
}

func waitForHttpServer(t *testing.T) {
	t.Helper()

	require.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			resp, err := http.Get("http://localhost:8080/health")
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			if assert.Less(t, resp.StatusCode, 300, "API not ready, http status: %d", resp.StatusCode) {
				return
			}
		},
		time.Second*10,
		time.Millisecond*50,
	)
}

func sendTicketsStatus(t *testing.T, req ticketsHttp.TicketStatusRequest) {
	t.Helper()

	requestBody := struct {
		Tickets []ticketsHttp.TicketStatusRequest `json:"tickets"`
	}{
		Tickets: []ticketsHttp.TicketStatusRequest{req},
	}

	payload, err := json.Marshal(requestBody)
	require.NoError(t, err)

	correlationID := shortuuid.New()
	idempotencyKey := shortuuid.New()

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertReceiptForTicketIssued(t *testing.T, receiptsService *adapters.ReceiptsServiceClientStub, ticket ticketsHttp.TicketStatusRequest) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			issuedReceipts := len(receiptsService.IssuedReceipts)
			parentT.Log("issued receipts", issuedReceipts)

			assert.Greater(t, issuedReceipts, 0, "no receipts issued")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var receipt entities.TicketBookingConfirmed
	var ok bool
	for _, issuedReceipt := range receiptsService.IssuedReceipts {
		if issuedReceipt.TicketID != ticket.TicketID {
			continue
		}
		receipt = issuedReceipt
		ok = true
		break
	}

	require.Truef(t, ok, "receipt for ticket %s not found", ticket.TicketID)
	assert.Equal(t, ticket.TicketID, receipt.TicketID)
	assert.Equal(t, ticket.Price.Amount, receipt.Price.Amount)
	assert.Equal(t, ticket.Price.Currency, receipt.Price.Currency)
}

func assertRowAddedToSpreadsheet(t *testing.T, spreadsheetsAPI *adapters.SpreadsheetsAPIClientStub, ticket ticketsHttp.TicketStatusRequest, expectedSpreadsheet string) {
	t.Helper()

	parentT := t

	assert.EventuallyWithT(
		t,
		func(t *assert.CollectT) {
			appendedRows := len(spreadsheetsAPI.AppendedRows)
			parentT.Log("appended rows", appendedRows)

			assert.Greater(t, appendedRows, 0, "no rows appended")
		},
		10*time.Second,
		100*time.Millisecond,
	)

	var foundRow []string
	var ok bool
	for _, row := range spreadsheetsAPI.AppendedRows {
		if len(row.Columns) > 0 && row.Columns[0] == ticket.TicketID {
			foundRow = row.Columns
			ok = true
			break
		}
	}

	require.Truef(t, ok, "row for ticket %s not found in %s spreadsheet", ticket.TicketID, expectedSpreadsheet)
	require.Len(t, foundRow, 4, "row should have 4 columns")
	assert.Equal(t, ticket.TicketID, foundRow[0])
	assert.Equal(t, ticket.CustomerEmail, foundRow[1])
	assert.Equal(t, ticket.Price.Amount, foundRow[2])
	assert.Equal(t, ticket.Price.Currency, foundRow[3])
}

func createShow(t *testing.T, showID string, numberOfTickets int) {
	t.Helper()

	show := ticketsHttp.ShowCreationRequest{
		ShowID:          showID,
		DeadNationID:    "dead-nation-" + showID,
		NumberOfTickets: numberOfTickets,
		StartTime:       time.Now().Add(24 * time.Hour),
		Title:           "Test Concert",
		Venue:           "Test Venue",
	}

	payload, err := json.Marshal(show)
	require.NoError(t, err)

	resp, err := http.Post(
		"http://localhost:8080/shows",
		"application/json",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func bookTickets(t *testing.T, showID string, numberOfTickets int, customerEmail string) *http.Response {
	t.Helper()

	booking := ticketsHttp.BookTicketsRequest{
		ShowID:          showID,
		NumberOfTickets: numberOfTickets,
		CustomerEmail:   customerEmail,
	}

	payload, err := json.Marshal(booking)
	require.NoError(t, err)

	resp, err := http.Post(
		"http://localhost:8080/book-tickets",
		"application/json",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	return resp
}

func assertBookingSucceeds(t *testing.T, resp *http.Response) {
	t.Helper()
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response ticketsHttp.BookTicketsResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.NotEmpty(t, response.BookingID, "booking ID should not be empty")
}

func assertBookingFails(t *testing.T, resp *http.Response, expectedMessage string) {
	t.Helper()
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errorResponse map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&errorResponse)
	require.NoError(t, err)

	message, ok := errorResponse["message"].(string)
	require.True(t, ok, "error response should have 'message' field")
	require.Contains(t, message, expectedMessage, "error message should mention insufficient seats")
}
