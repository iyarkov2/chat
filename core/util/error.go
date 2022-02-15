package util

/*
		Just an experiment with errors
 */
import "fmt"

type Error struct {
	Code uint
	Msg string
	Cause error
}

func (e Error) Error() string {
	if e.Cause == nil {
		return e.Msg
	} else {
		return fmt.Sprintf("%s caused by %s", e.Msg, e.Cause)
	}
}

func (e Error) Unwrap() error {
	return e.Cause
}

func (e Error) Is(target error) bool {
	if targetError, ok := target.(Error) ; ok {
		return targetError.Code == e.Code
	}
	return false
}

func (e Error) As(target interface{}) bool {
	if targetError, ok := target.(*Error) ; ok {
		targetError.Code = e.Code
		targetError.Msg = e.Msg
		targetError.Cause = e.Cause
		return true
	}
	return false
}


