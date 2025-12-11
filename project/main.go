package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"tickets/adapters"
	"tickets/database"
	"tickets/service"
	"tickets/tracing"
)

func main() {
	log.Init(slog.LevelInfo)

	tp, err := tracing.ConfigureTraceProvider()
	if err != nil {
		panic(err)
	}

	traceHttpClient := &http.Client{Transport: otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("HTTP %s %s %s", r.Method, r.URL.String(), operation)
		}),
	)}

	apiClients, err := clients.NewClientsWithHttpClient(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
		traceHttpClient,
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

	svc.SetTraceProvider(tp)

	err = svc.RunWithGracefulShutdown("")
	if err != nil {
		panic(err)
	}
}
