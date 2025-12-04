package adapters

import (
	"context"
	"net/http"
	"sync"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/dead_nation"
)

type DeadNationAPIStub struct {
	mu sync.Mutex

	BookingCalls []dead_nation.PostTicketBookingJSONRequestBody
}

func (d *DeadNationAPIStub) PostTicketBookingWithResponse(
	ctx context.Context,
	body dead_nation.PostTicketBookingJSONRequestBody,
	reqEditors ...dead_nation.RequestEditorFn,
) (*dead_nation.PostTicketBookingResponse, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.BookingCalls = append(d.BookingCalls, body)

	return &dead_nation.PostTicketBookingResponse{
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
	}, nil
}
