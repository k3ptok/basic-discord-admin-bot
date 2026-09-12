package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)
func InitLogger() *slog.Logger {
// Configure file rotation via Lumberjack
	fileLogger := &lumberjack.Logger{
		Filename:   "logs/bot.log", // Path to log file
		MaxSize:    10,             // Max megabytes before rotating
		MaxBackups: 5,              // Max number of old log files to keep
		MaxAge:     28,             // Max days to keep old log files
		Compress:   true,           // Compress rotated log files (.gz)
	}

	// Combine stdout (terminal) and fileLogger (disk)
	multiWriter := io.MultiWriter(os.Stdout, fileLogger)

	// Create JSON handler (Standard format for cloud hosts)
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Set minimum log level (Debug, Info, Warn, Error)
	})

	logger := slog.New(handler)
	
	// Set as Go's default global logger
	slog.SetDefault(logger)

	return logger
}