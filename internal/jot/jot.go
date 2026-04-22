package jot

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

// Use strings so the Go linker (-X) can modify them at compile time.
// go build -ldflags="-X 'website/internal/jot.sensitiveBuild=true' -X 'website/internal/jot.logFormat=gcp'" .
var (
	sensitiveBuild = "false"
	logFormat      = "text" // defaults to human-readable text for local dev
)

var internalLogger *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if logFormat == "gcp" {
		// --- GCP Cloud Run Mode (JSON with specific keys) ---
		opts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				a.Key = "severity"
			} else if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		}
		internalLogger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
	} else if logFormat == "json" {
		// --- Standard JSON Mode (No GCP quirks) ---
		internalLogger = slog.New(slog.NewJSONHandler(os.Stderr, opts))
	} else {
		// --- Local Dev Mode (Human-readable text) ---
		internalLogger = slog.New(slog.NewTextHandler(os.Stderr, opts))
	}
}

func SetLogger(l *slog.Logger) {
	internalLogger = l
}

func getCallerAttr(skip int) slog.Attr {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return slog.Attr{}
	}
	return slog.String("caller", fmt.Sprintf("%s:%d", filepath.Base(file), line))
}

// Log writes an error log if the error exists and returns true.
// Usage: if jot.Log(err, "failed to parse") { return }
func Log(e error, msg string, args ...any) bool {
	if e == nil {
		return false
	}

	if sensitiveBuild == "true" {
		// SENSITIVE PRODUCTION LOGIC: Ignore custom args
		internalLogger.Error(msg,
			slog.String("error", e.Error()),
			slog.String("errorType", fmt.Sprintf("%T", e)),
			getCallerAttr(2),
		)
	} else {
		// DEFAULT LOGIC
		allArgs := []any{
			slog.String("error", e.Error()),
			slog.String("errorType", fmt.Sprintf("%T", e)),
			getCallerAttr(2),
		}
		allArgs = append(allArgs, args...)
		internalLogger.Error(msg, allArgs...)
	}
	return true
}

// Info writes a standard informational log.
func Info(msg string, args ...any) {
	if sensitiveBuild == "true" {
		internalLogger.Info(msg, getCallerAttr(2))
	} else {
		allArgs := []any{getCallerAttr(2)}
		allArgs = append(allArgs, args...)
		internalLogger.Info(msg, allArgs...)
	}
}

// Panic logs the error and triggers a panic if the error exists.
func Panic(err error, msg string, args ...any) {
	if Log(err, msg, args...) {
		panic(err)
	}
}

// Fatal logs the error and immediately exits the application if the error exists.
func Fatal(err error, msg string, args ...any) {
	if Log(err, msg, args...) {
		os.Exit(1)
	}
}
