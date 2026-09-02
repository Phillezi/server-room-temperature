//go:build ignore

package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/phillezi/gob"
)

func init() {
	logger := slog.New(gob.NewPrettyHandler(nil))
	slog.SetDefault(logger)
}

func tinyGoBuild() gob.Target {
	return gob.NewCMD(
		func(ctx context.Context, cfg gob.Config) error {
			if err := os.MkdirAll(cfg.OutDir, os.ModePerm); err != nil {
				return fmt.Errorf("create output directory %q: %w", cfg.OutDir, err)
			}

			tinygoPath, err := exec.LookPath("tinygo")
			if err != nil {
				return fmt.Errorf(
					"tinygo not found in PATH: please install TinyGo from https://tinygo.org/getting-started/install/",
				)
			}

			out := cfg.OutDir + "/srt-sensor.uf2"
			var stderr bytes.Buffer

			cmd := exec.CommandContext(
				ctx,
				tinygoPath, "build",
				"-target", "pico",
				"-tags", "rp2040",
				"-o", out,
				"./cmd/srt-sensor/",
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				return fmt.Errorf(
					"tinygo build failed: %w\n%s\nhint: verify that tinygo supports your installed Go version and that the pico target is available",
					err,
					stderr.String(),
				)
			}

			slog.Default().Info("built sensor firmware", "output", out)
			return nil
		},
		"Build srt-sensor firmware with TinyGo for Raspberry Pi Pico => bin/srt-sensor.uf2",
	)
}

func allBuild() gob.Target {
	hostTarget := gob.Static()
	sensorTarget := tinyGoBuild()

	return gob.NewCMD(
		func(ctx context.Context, cfg gob.Config) error {
			slog.Default().Info("building host binaries...")
			if err := hostTarget.Run(ctx); err != nil {
				return fmt.Errorf("host build failed: %w", err)
			}

			slog.Default().Info("building sensor firmware...")
			if err := sensorTarget.Run(ctx); err != nil {
				return fmt.Errorf("sensor firmware build failed: %w", err)
			}

			return nil
		},
		"Build all binaries (host binaries and tinygo sensor firmware) => bin/",
	)
}

func main() {
	gob.New(gob.WithDefaultTarget("all")).
		Add("all", allBuild()).
		Add("host", gob.Static()).
		Add("sensor", tinyGoBuild()).
		Add("clean", gob.Clean()).
		Run()
}
