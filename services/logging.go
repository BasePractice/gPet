package services

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
)

func DefineLogging() *slog.Logger {
	var level *slog.Level
	debugLevel := slog.LevelDebug
	level = &debugLevel
	_ = level.UnmarshalText([]byte(LogLevel))
	opts := &slog.HandlerOptions{Level: level, AddSource: true}
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if LogFile != "" {
		file, err := os.OpenFile(LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			handler = slog.NewJSONHandler(file, opts)
		} else {
			log.Fatal(err)
		}
	} else if LogColor != "" {
		handler = tint.NewHandler(colorable.NewColorable(os.Stdout), &tint.Options{
			Level:      level,
			TimeFormat: time.DateTime,
			AddSource:  true,
		})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
