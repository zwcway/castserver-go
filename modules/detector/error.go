package detector

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/speaker"
)

type ResponseDecodeError struct {
	Field string
	Data  []byte
	Err   error
}

func (e *ResponseDecodeError) Error() string {
	return fmt.Sprintf("%s[%X]%s", e.Field, e.Data, e.Err)
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
	return fmt.Sprintf("unsupport speaker %d", e.Speaker.ID)
}
