package yip

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

// sesitiveBuild if true will not print out any arguments sent to yip.
// This variable will be set at compile time. Default is 'false'.
// go build -ldflags="-X 'website/pkg/yip.sensitiveBuild=true'" .
var sensitiveBuild = false

var internalLogger *slog.Logger

func init() {
	internalLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
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

// log an error if that error exists
func If(e error, msg string, args ...any) bool {
	if e == nil {
		return false
	}

	if sensitiveBuild {
		// The compiler will only include this block when sensitiveBuild is true.
		internalLogger.Error(msg,
			slog.String("error", e.Error()),
			slog.String("errorType", fmt.Sprintf("%T", e)),
			getCallerAttr(2),
		)
	} else {
		// This block gets compiled out when sensitiveBuild is true.
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

func Info(msg string, args ...any) {
	if sensitiveBuild {
		// SENSITIVE PRODUCTION LOGIC
		// Log the message and caller, but IGNORE the custom 'args' to prevent data leaks.
		internalLogger.Info(msg, getCallerAttr(2))
	} else {
		// DEFAULT DEVELOPMENT LOGIC
		// This block is only reached when sensitiveBuild is false.
		allArgs := []any{getCallerAttr(2)}
		allArgs = append(allArgs, args...)
		internalLogger.Info(msg, allArgs...)
	}
}

func Error(e error, msg string, args ...any) {
	If(e, msg, args...)
}

func Panic(err error, msg string, args ...any) {
	if If(err, msg, args...) {
		panic(err)
	}
}

func Fatal(msg string, args ...any) {
	if sensitiveBuild {
		internalLogger.Error(msg, getCallerAttr(2))
	} else {
		allArgs := []any{getCallerAttr(2)}
		allArgs = append(allArgs, args...)
		internalLogger.Error(msg, allArgs...)
	}
	os.Exit(1)
}
