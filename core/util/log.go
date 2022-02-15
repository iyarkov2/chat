package util

import (
	"context"
	"github.com/rs/zerolog"
	"os"
)

type logKeyType string
const logKey = logKeyType("util.log")

var parentLogger = zerolog.New(zerolog.ConsoleWriter {Out: os.Stdout, TimeFormat: "2006-01-02T15:04:0543"}).With().Timestamp().Logger()

func WithLogger(ctx context.Context, fields map[string]string) context.Context {
	logContext := parentLogger.With()

	for key, value := range fields {
		logContext = logContext.Str(key, value)
	}
	return context.WithValue(ctx, logKey, logContext.Logger())
}

func GetLogger(ctx context.Context) zerolog.Logger {
	val := ctx.Value(logKey)
	if val == nil {
		return parentLogger
	} else if log, ok := val.(zerolog.Logger); ok {
		return log
	} else {
		panic("Invalid logger in context")
	}
}

