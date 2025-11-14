package http

import (
	"net/http"
	"tickets/worker"
	"github.com/labstack/echo/v4"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request ticketsConfirmationRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		taskMessage := worker.Message{
			TicketID: ticket,
			Task:     worker.TaskIssueReceipt,
		}

		trackerMessage := worker.Message{
			TicketID: ticket,
			Task:     worker.TaskAppendToTracker,
		}

		h.Worker.Send(taskMessage, trackerMessage)
 	
	}

	return c.NoContent(http.StatusOK)
}
