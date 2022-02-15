package util

import (
	"context"
	"reflect"
)

type Closable interface {
	Close() error
}

func CloseQuiet(ctx context.Context, name string, closable Closable) {
	if closable == nil {
		return
	}
	if reflect.TypeOf(closable).Kind() == reflect.Ptr && reflect.ValueOf(closable).IsNil() {
		return
	}
	closeError := closable.Close()
	if closeError != nil {
		logger := GetLogger(ctx)
		logger.Error().Msgf("Error while closing %s %w", name, closeError)
	}
}

func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}