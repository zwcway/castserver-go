package speaker

import "github.com/zwcway/castserver-go/common/audio"

type SpeakerLineID uint8
type SpeakerLine struct {
	id   SpeakerLineID
	name string
}

type SpeakerModel uint8

const (
	SpeakerModel_UNICAST   SpeakerModel = 0
	SpeakerModel_MULTICAST SpeakerModel = 1
)

type SpeakerState uint32

const (
	SpeakerState_OFFLINE    SpeakerState = 0
	SpeakerState_ONLINE     SpeakerState = 0x00000001
	SpeakerState_CONNERROR  SpeakerState = 0x01000000
	SpeakerState_QUEUEERROR SpeakerState = 0x02000000
	SpeakerState_DELETED    SpeakerState = 0x80000000
)

type SpeakerID uint32

func (id *SpeakerID) IsValid() bool {
	return true
}

type SpeakerStatistic struct {
	Queue uint32
	Spend uint32
	Drop  uint32
}

var DefaultLine SpeakerLineID = 0
var DefaultChannel audio.AudioChannel = audio.AudioChannel_FRONT_LEFT
