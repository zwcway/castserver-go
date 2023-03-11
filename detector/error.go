package detector

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/speaker"
)

type UnpackError struct {
	Field string
	Data  []byte
	Err   error
}

func (e *UnpackError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s[%X]%s", e.Field, e.Data, e.Err.Error())
	}
	return fmt.Sprintf("%s[%X]", e.Field, e.Data)
}
func newUnpackError(f string, v []byte, err error) *UnpackError {
	return &UnpackError{f, v, err}
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

type UnsupportError struct {
	Speaker *speaker.Speaker
}

func (e *UnsupportError) Error() string {
	return fmt.Sprintf("unsupport speaker %d", e.Speaker.Id)
}
