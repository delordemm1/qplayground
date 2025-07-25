package platform

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func InitLogger() {
	w := os.Stderr
	var log slog.Level
	switch ENV_LOG_LEVEL {
	case "error":
		log = slog.LevelError
	case "warn":
		log = slog.LevelWarn
	case "info":
		log = slog.LevelInfo
	default:
		log = slog.LevelDebug
	}
	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			Level:      log,
			TimeFormat: time.Kitchen,
		}),
	))
}

func Perf(msg string, start time.Time) {
	slog.Info(msg, "duration", time.Since(start))
}
