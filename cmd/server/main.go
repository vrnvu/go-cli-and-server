package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jaevor/go-nanoid"
	"github.com/vrnvu/temp/internal/server"
)

func fromEnvPort() string {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	return port
}

func fromEnvSlog() (*slog.Logger, error) {
	logLevel := slog.LevelInfo
	if v, ok := os.LookupEnv("LOG_LEVEL"); ok {
		switch v {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			return nil, fmt.Errorf("invalid log level: `%s`, try: [debug, info, warn, error]", v)
		}
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})), nil
}

// TODO tls/https
func main() {
	port := fromEnvPort()
	slog, err := fromEnvSlog()
	if err != nil {
		panic(err)
	}

	requestIDGenerator, err := nanoid.Canonic()
	if err != nil {
		panic(err)
	}

	quizHandler, err := server.FromConfig(&server.Config{
		Slog:               slog,
		RequestIDGenerator: requestIDGenerator,
	})
	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           quizHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		slog.Info("starting server", "port", port)
		if err := server.ListenAndServeTLS("localhost.pem", "localhost-key.pem"); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited properly")
}
