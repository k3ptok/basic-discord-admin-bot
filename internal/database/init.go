package database

import (
	"database/sql"
	"embed"
	"log/slog"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func InitDB(logger *slog.Logger, migrations embed.FS) *sql.DB {
	db, err := sql.Open("sqlite", "data/bot.db?_pragma=journal_mode(wal)&_pragma=busy_timeout(5000)")
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		panic(err)
	}

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		logger.Error("Failed to set goose dialect", "error", err)
		panic(err)
	}

	if err := goose.Up(db, "sql/schema"); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		panic(err)
	}

	logger.Info("Database migrations applied successfully")
	return db
}