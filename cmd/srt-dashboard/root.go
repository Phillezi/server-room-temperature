package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/phillezi/server-room-temperature/frontend"
	"github.com/phillezi/server-room-temperature/internal/config"
	"github.com/phillezi/server-room-temperature/internal/defaults"
	"github.com/phillezi/server-room-temperature/internal/history"
	"github.com/phillezi/server-room-temperature/internal/middleware"
	"github.com/phillezi/server-room-temperature/internal/natsconn"
)

var rootCmd = cobra.Command{
	Use:     use,
	Short:   short,
	Long:    long,
	Version: version,
	RunE: func(cmd *cobra.Command, _ []string) error {
		natsCfg := natsconn.Config{
			URL:      viper.GetString("nats.url"),
			Host:     viper.GetString("nats.host"),
			User:     viper.GetString("nats.user"),
			Password: viper.GetString("nats.password"),
		}
		streamName := viper.GetString("stream")
		httpAddr := viper.GetString("http.addr")
		if httpAddr != "" && httpAddr[0] != ':' && !strings.Contains(httpAddr, ":") {
			httpAddr = ":" + httpAddr
		}

		frontendCfg := frontend.Config{
			NatsWSURL:    viper.GetString("nats.ws_url"),
			NatsUser:     viper.GetString("nats.reader_user"),
			NatsPassword: viper.GetString("nats.reader_password"),
			Subject:      viper.GetString("nats.subject"),
		}

		var (
			nc     *nats.Conn
			js     jetstream.JetStream
			stream jetstream.Stream
		)

		for {
			if nc == nil {
				var err error
				nc, err = natsconn.Connect(natsCfg)
				if err != nil {
					log.Printf("waiting for NATS connection (%s): %v", natsconn.FormatURL(natsCfg), err)
					select {
					case <-cmd.Context().Done():
						return cmd.Context().Err()
					case <-time.After(2 * time.Second):
						continue
					}
				}
			}

			if js == nil {
				var err error
				js, err = jetstream.New(nc)
				if err != nil {
					log.Printf("waiting for JetStream context: %v", err)
					select {
					case <-cmd.Context().Done():
						nc.Close()
						return cmd.Context().Err()
					case <-time.After(2 * time.Second):
						continue
					}
				}
			}

			var err error
			stream, err = js.Stream(cmd.Context(), streamName)
			if err != nil {
				log.Printf("waiting for JetStream stream %q to be ready: %v", streamName, err)
				select {
				case <-cmd.Context().Done():
					nc.Close()
					return cmd.Context().Err()
				case <-time.After(2 * time.Second):
					continue
				}
			}

			break
		}
		defer nc.Close()

		historyService := history.New(stream)

		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		mux.Handle("/api/history", history.HistoryHandler{Service: historyService})
		mux.Handle("/", frontend.Handler(frontendCfg))

		server := &http.Server{
			Addr:    httpAddr,
			Handler: middleware.CORS(mux),
		}

		go func() {
			log.Printf("history API listening on %s", server.Addr)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("http listen: %v", err)
			}
		}()

		<-cmd.Context().Done()
		log.Println("shutting down dashboard server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	},
}

func init() {
	cobra.OnInitialize(func() { config.Init(use) })

	rootCmd.Flags().String("nats-url", "", "NATS server URL (e.g. nats://localhost:4222)")
	viper.BindPFlag("nats.url", rootCmd.Flags().Lookup("nats-url"))

	rootCmd.Flags().String("nats-host", defaults.DefaultNATSHost, "NATS server host")
	viper.BindPFlag("nats.host", rootCmd.Flags().Lookup("nats-host"))

	rootCmd.Flags().String("nats-user", defaults.DefaultNATSDashboardUser, "NATS username for backend API")
	viper.BindPFlag("nats.user", rootCmd.Flags().Lookup("nats-user"))

	rootCmd.Flags().String("nats-password", defaults.DefaultNATSDashboardPassword, "NATS password for backend API")
	viper.BindPFlag("nats.password", rootCmd.Flags().Lookup("nats-password"))

	rootCmd.Flags().String("nats-ws-url", defaults.DefaultNATSWSURL, "NATS WebSocket URL for browser frontend")
	viper.BindPFlag("nats.ws_url", rootCmd.Flags().Lookup("nats-ws-url"))

	rootCmd.Flags().String("nats-reader-user", defaults.DefaultNATSReaderUser, "NATS username for browser frontend")
	viper.BindPFlag("nats.reader_user", rootCmd.Flags().Lookup("nats-reader-user"))

	rootCmd.Flags().String("nats-reader-password", defaults.DefaultNATSReaderPassword, "NATS password for browser frontend")
	viper.BindPFlag("nats.reader_password", rootCmd.Flags().Lookup("nats-reader-password"))

	rootCmd.Flags().StringP("topic", "t", defaults.DefaultNATSTopic, "NATS default topic/subject for frontend")
	viper.BindPFlag("nats.subject", rootCmd.Flags().Lookup("topic"))

	rootCmd.Flags().StringP("stream", "s", defaults.DefaultNATSStream, "JetStream stream name")
	viper.BindPFlag("stream", rootCmd.Flags().Lookup("stream"))

	rootCmd.Flags().StringP("http-addr", "a", defaults.DefaultHTTPAddr, "HTTP server address to listen on")
	viper.BindPFlag("http.addr", rootCmd.Flags().Lookup("http-addr"))

	viper.BindEnv("nats.url", "NATS_URL")
	viper.BindEnv("nats.host", "NATS_HOST")
	viper.BindEnv("nats.user", "NATS_USER")
	viper.BindEnv("nats.password", "NATS_PASSWORD")
	viper.BindEnv("nats.ws_url", "NATS_WS_URL")
	viper.BindEnv("nats.reader_user", "NATS_READER_USER")
	viper.BindEnv("nats.reader_password", "NATS_READER_PASSWORD")
	viper.BindEnv("nats.subject", "NATS_SUBJECT", "NATS_TOPIC")
	viper.BindEnv("stream", "STREAM_NAME")
	viper.BindEnv("http.addr", "HTTP_ADDR", "PORT")
}
