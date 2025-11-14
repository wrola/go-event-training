package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type ticketsConfirmationRequest struct {
	Tickets []string `json:"tickets"`
}

func (h Handler) PostTicketsConfirmation(c echo.Context) error {
	var request ticketsConfirmationRequest
	if err := c.Bind(&request);  err != nil {
		return err
	}

	for _, ticket := range request.Tickets {
		taskMessage := message.NewMessage(watermill.NewUUID(), []byte(ticket))
		trackerMessage := message.NewMessage(watermill.NewUUID(), []byte(ticket))

		h.publisher.Publish("issue-receipt", taskMessage)
		h.publisher.Publish("append-to-tracker", trackerMessage)
	}

	return c.NoContent(http.StatusOK)
}
