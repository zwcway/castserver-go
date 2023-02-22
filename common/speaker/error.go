package speaker

import (
	"fmt"
	"net"
)

type LineError struct {
	Line Line
	Err  error
}

func (e *LineError) Error() string {
	return fmt.Sprintf("The line %s(%d) got error: %s", e.Line.Name, e.Line.ID, e.Err)
}

type UnknownLineError struct {
	Line LineID
}

func (e *UnknownLineError) Error() string {
	return fmt.Sprintf("The line %d is invalid", e.Line)
}

type SpeakerError struct {
	ID   ID
	Name string
	IP   net.IP
	MAC  net.HardwareAddr
	Err  error
}

func (e *SpeakerError) Error() string {
	return fmt.Sprintf("The speaker %d got error: %s", e.ID, e.Err)
}

type UnknownSpeakerError struct {
	ID ID
}

func (e *UnknownSpeakerError) Error() string {
	return fmt.Sprintf("The speaker %d is invalid", e.ID)
}
