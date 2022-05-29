package handlers

import "fmt"

// UniqueIDViolatedError ошибка позволяющая передавать map[id]actual_id для последующей замены.
// map делал на вырост, чтобы сделать обновление для пакетной обработки. Но не осилил из-за CorrelationID
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
