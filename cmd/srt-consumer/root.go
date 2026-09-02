package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/phillezi/server-room-temperature/internal/config"
	"github.com/phillezi/server-room-temperature/internal/defaults"
	"github.com/phillezi/server-room-temperature/internal/dto"
	"github.com/phillezi/server-room-temperature/internal/natsconn"
	"github.com/phillezi/server-room-temperature/pkg/protocol"
)

var rootCmd = cobra.Command{
	Use:     use,
	Short:   short,
	Long:    long,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		device := viper.GetString("device")
		if len(args) > 0 && args[0] != "" {
			device = args[0]
		}
		if device == "" {
			return fmt.Errorf("serial device required (pass as argument or via --device / SERIAL_PORT)")
		}

		natsCfg := natsconn.Config{
			URL:      viper.GetString("nats.url"),
			Host:     viper.GetString("nats.host"),
			User:     viper.GetString("nats.user"),
			Password: viper.GetString("nats.password"),
		}
		topic := viper.GetString("nats.topic")

		f, err := os.Open(device)
		if err != nil {
			return fmt.Errorf("open serial device %s: %w", device, err)
		}
		defer f.Close()

		nc, err := natsconn.Connect(natsCfg)
		if err != nil {
			return err
		}
		defer nc.Close()

		fr := NewFrameReader(f)
		var frame [maxFrameSize]byte

		go func() {
			<-cmd.Context().Done()
			f.Close()
		}()

		for {
			n, err := fr.ReadFrame(frame[:])
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed) || cmd.Context().Err() != nil {
					return nil
				}
				return fmt.Errorf("read frame: %w", err)
			}

			milliC, ok := protocol.Decode(frame[:n])
			if !ok {
				continue
			}

			reading := dto.Reading{
				Timestamp:  time.Now().UTC(),
				TempMilliC: milliC,
			}

			data, err := json.Marshal(reading)
			if err != nil {
				log.Printf("marshal reading: %v", err)
				continue
			}

			if err := nc.Publish(topic, data); err != nil {
				log.Printf("publish: %v", err)
				continue
			}
		}
	},
}

func init() {
	cobra.OnInitialize(func() { config.Init(use) })

	rootCmd.Flags().String("nats-url", "", "NATS server URL (e.g. nats://localhost:4222)")
	viper.BindPFlag("nats.url", rootCmd.Flags().Lookup("nats-url"))

	rootCmd.Flags().String("nats-host", defaults.DefaultNATSHost, "NATS server host")
	viper.BindPFlag("nats.host", rootCmd.Flags().Lookup("nats-host"))

	rootCmd.Flags().String("nats-user", defaults.DefaultNATSWriterUser, "NATS username")
	viper.BindPFlag("nats.user", rootCmd.Flags().Lookup("nats-user"))

	rootCmd.Flags().String("nats-password", defaults.DefaultNATSWriterPassword, "NATS password")
	viper.BindPFlag("nats.password", rootCmd.Flags().Lookup("nats-password"))

	rootCmd.Flags().StringP("topic", "t", defaults.DefaultNATSTopic, "NATS topic/subject to publish to")
	viper.BindPFlag("nats.topic", rootCmd.Flags().Lookup("topic"))

	rootCmd.Flags().StringP("device", "d", "", "Serial device path (e.g. /dev/ttyACM0)")
	viper.BindPFlag("device", rootCmd.Flags().Lookup("device"))

	viper.BindEnv("nats.url", "NATS_URL")
	viper.BindEnv("nats.host", "NATS_HOST")
	viper.BindEnv("nats.user", "NATS_USER")
	viper.BindEnv("nats.password", "NATS_PASSWORD")
	viper.BindEnv("nats.topic", "NATS_TOPIC", "NATS_SUBJECT")
	viper.BindEnv("device", "SERIAL_PORT", "DEVICE")
}
