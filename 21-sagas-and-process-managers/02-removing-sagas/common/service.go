package common

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/v2/common/http"
	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

type AddHandlersFn func(e *echo.Echo)

func StartService(ctx context.Context, addHandlers []AddHandlersFn) {
	log.Init(slog.LevelInfo)

	e := commonHTTP.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	for _, addHandlerFn := range addHandlers {
		addHandlerFn(e)
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	errgrp := errgroup.Group{}

	errgrp.Go(func() error {
		return e.Start(":8080")
	})

	errgrp.Go(func() error {
		<-ctx.Done()
		return e.Shutdown(context.Background())
	})

	if err := errgrp.Wait(); err != nil {
		panic(err)
	}
}
