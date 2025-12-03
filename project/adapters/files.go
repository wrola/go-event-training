package adapters

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
)

type FilesAPIClientStub struct {
	SaveFiles []string
	lock         sync.Mutex
}

func (f *FilesAPIClientStub) StoreFile(ctx context.Context, fileID string, fileContent string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.SaveFiles = append(f.SaveFiles, fileID) 

	return nil
}

type FilesAPIClient struct {
	clients *clients.Clients
}

func NewFilesAPIClient(clients *clients.Clients) *FilesAPIClient {
	if clients == nil {
		panic("NewFilesAPIClient: clients is nil")
	}
	return &FilesAPIClient{clients: clients}
}

func (c FilesAPIClient) StoreFile(ctx context.Context, fileID string, fileContent string) error {
	resp, err := c.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(
		ctx,
		fileID,
		fileContent,
	)
	if err != nil {
		return fmt.Errorf("failed to store file: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusCreated:
		return nil
	case http.StatusConflict:
		log.FromContext(ctx).With("file", fileID).Info("file already exists")
		return nil
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
}
