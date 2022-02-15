package util

import (
	"errors"
	"fmt"
	"testing"
)

func Test_Construct(t *testing.T) {
	err := Error {
		Code : 1,
		Msg : "First Error",
	}
	if err.Code != 1 {
		t.Errorf("code expected to be 1")
	}
	if err.Msg != "First Error" {
		t.Errorf("msg expected to be nil")
	}
	if err.Cause != nil {
		t.Errorf("Cause expected to be nil")
	}
}

func Test_Format(t *testing.T) {
	err := Error {
		Code : 1,
		Msg : "First Error",
	}
	formatted := fmt.Errorf("Error: [%w]", err).Error()
	expected := "Error: [First Error]"
	if formatted != expected  {
		t.Errorf("Inval message, expected [%s], actual: [%s]", expected, formatted)
	}
}

func Test_FormatWithCause(t *testing.T) {
	err := Error {
		Code : 1,
		Msg : "First Error",
		Cause: errors.New("some Nested Error"),
	}
	formatted := fmt.Errorf("Error: [%w]", err).Error()
	expected := "Error: [First Error caused by some Nested Error]"
	if formatted != expected  {
		t.Errorf("Inval message, expected [%s], actual: [%s]", expected, formatted)
	}
}

func Test_Wrap(t *testing.T) {
	err := Error {
		Code : 1,
		Msg : "First Error",
	}
	wrapped := fmt.Errorf("Error: [%w]", err)
	unwrapped := errors.Unwrap(wrapped)
	if err != unwrapped {
		t.Errorf("Error does not work with fmt.Errorf/errors.Unwrap methods as expected")
	}
}

func Test_Unwrap(t *testing.T) {
	cause := errors.New("some error")
	err := Error {
		Code : 1,
		Msg : "First Error",
		Cause : cause,
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Error does not work with errors.Unwrap methods as expected")
	}
}


func Test_IsPositive(t *testing.T) {
	err1 := Error {
		Code : 23,
		Msg : "First Error",
	}
	err2 := Error {
		Code : 23,
		Msg : "Second Error",
	}

	if !errors.Is(err1, err1) {
		t.Errorf("errors.Is(err1, err1) expected to be true")
	}

	if !errors.Is(err1, err2) {
		t.Errorf("errors.Is(err1, err2) expected to be true")
	}

	if !errors.Is(err2, err1) {
		t.Errorf("errors.Is(err1, err2) expected to be true")
	}
}

func Test_IsPositiveInChain(t *testing.T) {
	err := fmt.Errorf("top Level Error: [%w]", fmt.Errorf("error: [%w]", Error {
		Code : 23,
		Msg : "First Error",
		Cause : errors.New("some error"),
	}))
	target := Error { Code : 23 }

	if !errors.Is(err, target) {
		t.Errorf("errors.Is(err, target) expected to be true")
	}
}

func Test_AsInChain(t *testing.T) {
	cause := errors.New("some error")
	err := fmt.Errorf("top Level Error: [%w]", fmt.Errorf("error: [%w]", Error {
		Code : 23,
		Msg : "First Error",
		Cause : cause,
	}))
	target := Error {}
	errors.As(err, &target)
	if target.Code != 23 {
		t.Errorf("target.Code expected to be 23")
	}
	if target.Msg != "First Error" {
		t.Errorf("target.Msg expected to be [First Error]")
	}
	if target.Cause != cause {
		t.Errorf("target.Cause expected to be the same as err.Cause")
	}
}

