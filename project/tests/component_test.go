package tests_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"tickets/adapters"
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

	go func() {
		svc, err := service.New(
			spreadsheetsAPI,
			receiptsService,
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

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://localhost:8080/tickets-status",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)

	httpReq.Header.Set("Correlation-ID", correlationID)
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
