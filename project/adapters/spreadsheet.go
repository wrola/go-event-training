package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients/spreadsheets"
)

type SpreadsheetsAPIClient struct {
	// we are not mocking this client: it's pointless to use interface here
	clients *clients.Clients
}

func NewSpreadsheetsAPIClient(clients *clients.Clients) *SpreadsheetsAPIClient {
	if clients == nil {
		panic("NewSpreadsheetsAPIClient: clients is nil")
	}

	return &SpreadsheetsAPIClient{clients: clients}
}

func (c SpreadsheetsAPIClient) AppendRow(ctx context.Context, spreadsheetName string, row []string) error {
	resp, err := c.clients.Spreadsheets.PostSheetsSheetRowsWithResponse(ctx, spreadsheetName, spreadsheets.PostSheetsSheetRowsJSONRequestBody{
		Columns: row,
	})
	if err != nil {
		return fmt.Errorf("failed to post row: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to post row: unexpected status code %d", resp.StatusCode())
	}

	return nil
}
