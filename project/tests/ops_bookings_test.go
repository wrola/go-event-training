package tests_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/lithammer/shortuuid/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tickets/entities/models"
	ticketsHttp "tickets/http"
)

func getOpsBookings(t *testing.T, queryParams string) ([]models.OpsBooking, int) {
	url := "http: localhost:8080/ops/bookings"
	if queryParams != "" {
		url += "?" + queryParams
	}

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var bookings []models.OpsBooking
	err = json.Unmarshal(body, &bookings)
	require.NoError(t, err)

	return bookings, resp.StatusCode
}

func getOpsBookingByID(t *testing.T, bookingID string) (*models.OpsBooking, int) {
	url := fmt.Sprintf("http: localhost:8080/ops/bookings/%s", bookingID)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var booking models.OpsBooking
	err = json.Unmarshal(body, &booking)
	require.NoError(t, err)

	return &booking, resp.StatusCode
}

func waitForOpsBooking(t *testing.T, bookingID string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, status := getOpsBookingByID(t, bookingID)
		if status == http.StatusOK {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("Ops booking %s not created within timeout", bookingID)
}

func extractBookingID(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response ticketsHttp.BookTicketsResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.NotEmpty(t, response.BookingID, "booking ID should not be empty")

	return response.BookingID
}

func confirmTickets(t *testing.T, ticketIDs []string, bookingID string) {
	t.Helper()

	for _, ticketID := range ticketIDs {
		req := ticketsHttp.TicketStatusRequest{
			TicketID:      ticketID,
			CustomerEmail: "test@example.com",
			Price: models.Money{
				Amount:   "100.00",
				Currency: "USD",
			},
			Status:    models.TicketStatusConfirmed,
			BookingID: bookingID,
		}
		sendTicketsStatus(t, req)
	}
}

func TestOpsBookingsEndpoints(t *testing.T) {
	t.Run("get all ops bookings without filter", func(t *testing.T) {
		showID := shortuuid.New()
		createShow(t, showID, 100)

		booking1Resp := bookTickets(t, showID, 1, "customer1@example.com")
		booking1ID := extractBookingID(t, booking1Resp)

		booking2Resp := bookTickets(t, showID, 1, "customer2@example.com")
		booking2ID := extractBookingID(t, booking2Resp)

		waitForOpsBooking(t, booking1ID, 5*time.Second)
		waitForOpsBooking(t, booking2ID, 5*time.Second)

		bookings, status := getOpsBookings(t, "")

		require.Equal(t, http.StatusOK, status)
		assert.GreaterOrEqual(t, len(bookings), 2, "Should contain at least our 2 test bookings")

		foundBooking1 := false
		foundBooking2 := false
		for _, b := range bookings {
			if b.BookingID.String() == booking1ID {
				foundBooking1 = true
			}
			if b.BookingID.String() == booking2ID {
				foundBooking2 = true
			}
		}
		assert.True(t, foundBooking1, "Booking 1 should be in results")
		assert.True(t, foundBooking2, "Booking 2 should be in results")
	})

	t.Run("get ops bookings filtered by receipt date", func(t *testing.T) {
		showID := shortuuid.New()
		createShow(t, showID, 100)

		bookingResp := bookTickets(t, showID, 2, "datetest@example.com")
		bookingID := extractBookingID(t, bookingResp)

		waitForOpsBooking(t, bookingID, 5*time.Second)

		booking, _ := getOpsBookingByID(t, bookingID)
		require.NotNil(t, booking)

		var ticketIDs []string
		for tid := range booking.Tickets {
			ticketIDs = append(ticketIDs, tid)
		}
		require.Equal(t, 2, len(ticketIDs))

		confirmTickets(t, ticketIDs, bookingID)

		time.Sleep(1 * time.Second)

		booking, _ = getOpsBookingByID(t, bookingID)

		var receiptDate string
		for _, ticket := range booking.Tickets {
			if !ticket.ReceiptIssuedAt.IsZero() {
				receiptDate = ticket.ReceiptIssuedAt.Format("2006-01-02")
				break
			}
		}
		require.NotEmpty(t, receiptDate, "Receipt should have been issued")

		bookings, status := getOpsBookings(t, fmt.Sprintf("receipt_issue_date=%s", receiptDate))

		require.Equal(t, http.StatusOK, status)
		assert.GreaterOrEqual(t, len(bookings), 1)

		found := false
		for _, b := range bookings {
			if b.BookingID.String() == bookingID {
				found = true
				break
			}
		}
		assert.True(t, found, "Booking should be returned when filtering by its receipt date")

		bookings, status = getOpsBookings(t, "receipt_issue_date=2020-01-01")
		require.Equal(t, http.StatusOK, status)

		found = false
		for _, b := range bookings {
			if b.BookingID.String() == bookingID {
				found = true
				break
			}
		}
		assert.False(t, found, "Booking should NOT be returned when filtering by different date")
	})

	t.Run("get ops booking by id - success", func(t *testing.T) {
		showID := shortuuid.New()
		createShow(t, showID, 100)

		bookingResp := bookTickets(t, showID, 1, "byid@example.com")
		bookingID := extractBookingID(t, bookingResp)

		waitForOpsBooking(t, bookingID, 5*time.Second)

		booking, status := getOpsBookingByID(t, bookingID)

		require.Equal(t, http.StatusOK, status)
		require.NotNil(t, booking)
		assert.Equal(t, bookingID, booking.BookingID.String())
		assert.NotNil(t, booking.Tickets)
	})

	t.Run("get ops booking by id - invalid uuid", func(t *testing.T) {
		url := "http: localhost:8080/ops/bookings/not-a-uuid"
		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var errorResp map[string]interface{}
		json.Unmarshal(body, &errorResp)

		assert.Contains(t, errorResp["message"], "not a valid UUID")
	})

	t.Run("invalid date format returns 400", func(t *testing.T) {
		testCases := []struct {
			name      string
			dateParam string
		}{
			{"invalid string", "invalid"},
			{"wrong format MM-DD-YYYY", "12-08-2025"},
			{"wrong format DD/MM/YYYY", "08/12/2025"},
			{"invalid date values", "2025-13-40"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				url := fmt.Sprintf("http: localhost:8080/ops/bookings?receipt_issue_date=%s", tc.dateParam)
				resp, err := http.Get(url)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

				body, _ := io.ReadAll(resp.Body)
				var errorResp map[string]interface{}
				json.Unmarshal(body, &errorResp)

				assert.Contains(t, errorResp["message"], "invalid date format")
				assert.Contains(t, errorResp["message"], "expected YYYY-MM-DD")
			})
		}
	})

	t.Run("empty date parameter returns all bookings", func(t *testing.T) {
		showID := shortuuid.New()
		createShow(t, showID, 100)

		bookingResp := bookTickets(t, showID, 1, "empty@example.com")
		bookingID := extractBookingID(t, bookingResp)
		waitForOpsBooking(t, bookingID, 5*time.Second)

		bookings, status := getOpsBookings(t, "receipt_issue_date=")

		require.Equal(t, http.StatusOK, status)
		assert.GreaterOrEqual(t, len(bookings), 1, "Should return bookings when date param is empty")

		found := false
		for _, b := range bookings {
			if b.BookingID.String() == bookingID {
				found = true
				break
			}
		}
		assert.True(t, found, "Booking should be returned when date filter is empty")
	})
}
