package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"

	"tickets/adapters"
	"tickets/database"
	"tickets/service"
)

func main() {
	log.Init(slog.LevelInfo)

	apiClients, err := clients.NewClients(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	spreadsheetsAPI := adapters.NewSpreadsheetsAPIClient(apiClients)
	receiptsService := adapters.NewReceiptsServiceClient(apiClients)
	paymentsService := adapters.NewPaymentsServiceClient(apiClients)
	filesAPI := adapters.NewFilesAPIClient(apiClients)
	deadNationAPI := apiClients.DeadNation

	db := database.NewDatabaseConnection()
	defer db.Close()

	svc, err := service.New(
		spreadsheetsAPI,
		receiptsService,
		paymentsService,
		filesAPI,
		deadNationAPI,
		db,
	)
	if err != nil {
		panic(err)
	}

	err = svc.RunWithGracefulShutdown("")
	if err != nil {
		panic(err)
	}
}
