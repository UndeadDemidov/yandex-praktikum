package handlers

import "fmt"

type UniqueIDViolatedError struct {
	ActualIDs map[string]string
	Err       error
}

func (e *UniqueIDViolatedError) Error() string {
	return fmt.Sprintf("%v: actual id is %v", e.Err, e.ActualIDs)
}

func NewUniqueIDViolatedError(err error, id map[string]string) error {
	return &UniqueIDViolatedError{
		ActualIDs: id,
		Err:       err,
	}
}

func (e *UniqueIDViolatedError) Unwrap() error { return e.Err }
