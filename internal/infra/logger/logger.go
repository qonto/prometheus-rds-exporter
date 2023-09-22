// Package logger implements logging methods
package logger

import (
	"log/slog"
	"os"
)

func New(debug bool, logFormat string) (*slog.Logger, error) {
	logLevel := &slog.LevelVar{}
	if debug {
		logLevel.Set(slog.LevelDebug)
	}

	opts := slog.HandlerOptions{
		Level: logLevel,
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	if logFormat == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	}

	return logger, nil
}
