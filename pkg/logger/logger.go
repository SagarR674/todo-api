// Package logger configures structured (JSON) application logging built on the
// standard library's log/slog. A single process-wide logger is initialised at
// startup and retrieved elsewhere via L().
package logger

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

var base *slog.Logger

// Init creates the JSON logger and installs it as the slog default.
// level is one of: debug, info, warn, error.
func Init(level string) {
	base = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	}))
	slog.SetDefault(base)
}

// L returns the process logger, initialising a default one if Init was skipped.
func L() *slog.Logger {
	if base == nil {
		Init("info")
	}
	return base
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GormLogger adapts GORM's logger interface onto slog so that database queries
// and errors land in the same structured log stream as the rest of the app.
type GormLogger struct {
	SlowThreshold time.Duration
	LogLevel      gormlogger.LogLevel
}

// NewGormLogger builds a GORM logger. In production only errors and slow
// queries are logged; in development all statements are logged at debug level.
func NewGormLogger(production bool) gormlogger.Interface {
	lvl := gormlogger.Info
	if production {
		lvl = gormlogger.Warn
	}
	return &GormLogger{SlowThreshold: 200 * time.Millisecond, LogLevel: lvl}
}

func (g *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *g
	clone.LogLevel = level
	return &clone
}

func (g *GormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= gormlogger.Info {
		L().Info("gorm: "+msg, slog.Any("data", data))
	}
}

func (g *GormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= gormlogger.Warn {
		L().Warn("gorm: "+msg, slog.Any("data", data))
	}
}

func (g *GormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if g.LogLevel >= gormlogger.Error {
		L().Error("gorm: "+msg, slog.Any("data", data))
	}
}

// Trace is called by GORM after every SQL statement.
func (g *GormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if g.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gormlogger.ErrRecordNotFound) && g.LogLevel >= gormlogger.Error:
		L().Error("db query failed",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
			slog.String("error", err.Error()),
		)
	case g.SlowThreshold > 0 && elapsed > g.SlowThreshold && g.LogLevel >= gormlogger.Warn:
		L().Warn("db slow query",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
		)
	case g.LogLevel >= gormlogger.Info:
		L().Debug("db query",
			slog.String("sql", sql),
			slog.Int64("rows", rows),
			slog.Duration("elapsed", elapsed),
		)
	}
}
