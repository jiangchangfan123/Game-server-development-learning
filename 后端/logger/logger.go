package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func Init() error {
	logDir := filepath.Join("..", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "server.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewTextHandler(multiWriter, opts)
	slog.SetDefault(slog.New(handler))

	return nil
}
