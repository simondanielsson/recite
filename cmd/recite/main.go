package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/simondanielsson/recite/cmd/internal/config"
	"github.com/simondanielsson/recite/cmd/internal/logging"
	"github.com/simondanielsson/recite/cmd/internal/server"
)

func main() {
	getenv := os.Getenv
	ctx := context.Background()

	if err := run(ctx, getenv, os.Stdout, os.Stderr, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(0)
	}
}

func run(ctx context.Context, getenv func(string) string, outWriter io.Writer, errWriter io.Writer, args []string) error {
	config, err := config.Load(getenv)
	if err != nil {
		return err
	}

	logger := logging.NewLogger(outWriter)
	server := server.New(config, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		// Note: goroutine waits here until interrupt
		<-quit
		logger.Println("shutting down server...")

		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("failed to shutdown server: %v", err)
		}
		logger.Println("server stopped gracefully")
	}()

	logger.Printf("listening on %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(errWriter, "error listening and serving: %s\n", err)
		return err
	}
	return nil
}
