package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/phillezi/server-room-temperature/frontend"
	"github.com/phillezi/server-room-temperature/internal/history"
)

func main() {
	nc, err := nats.Connect(
		"nats://temperature-dashboard:CHANGE-ME-DASHBOARD@localhost:4222",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nats connect: %v\n", err)
		os.Exit(1)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create jetstream: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	stream, err := js.Stream(ctx, "SERVER_ROOM_TEMPERATURE")
	if err != nil {
		fmt.Fprintf(os.Stderr, "get serverroom stream: %v\n", err)
		os.Exit(1)
	}

	historyService := history.New(stream)

	mux := http.NewServeMux()

	mux.Handle(
		"/api/history",
		history.HistoryHandler{
			Service: historyService,
		},
	)

	mux.Handle("/", frontend.Handler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("history API listening on %s", server.Addr)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
