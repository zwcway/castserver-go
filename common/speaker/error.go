package speaker

import (
	"fmt"
	"net"
)

type LineError struct {
	Line SpeakerLine
	Err  error
}

func (e *LineError) Error() string {
	return fmt.Sprintf("The line %s(%d) got error: %s", e.Line.name, e.Line.id, e.Err)
}

type UnknownLineError struct {
	Line SpeakerLineID
}

func (e *UnknownLineError) Error() string {
	return fmt.Sprintf("The line %d is invalid", e.Line)
}

type SpeakerError struct {
	ID   SpeakerID
	Name string
	IP   net.IP
	MAC  net.HardwareAddr
	Err  error
}

func (e *SpeakerError) Error() string {
	return fmt.Sprintf("The speaker %d got error: %s", e.ID, e.Err)
}

type UnknownSpeakerError struct {
	ID SpeakerID
}

func (e *UnknownSpeakerError) Error() string {
	return fmt.Sprintf("The speaker %d is invalid", e.ID)
}
