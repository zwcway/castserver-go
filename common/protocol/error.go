package protocol

import "fmt"

type Error struct {
	Desc string
}

func (e *Error) Error() string {
	return fmt.Sprintf("procotol invalid %s", e.Desc)
}

func NewError(desc string) *Error {
	return &Error{desc}
}
func NewOverError() *Error {
	return &Error{"package overflow"}
}
