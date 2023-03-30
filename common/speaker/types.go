package speaker

type LineID = uint8

const LineID_MAX LineID = (1 << 8) - 1

type SpeakerID = uint32

const SpeakerID_MAX SpeakerID = (1 << 32) - 1

var DefaultLineID LineID = 1
var maxLineID LineID = 1
var maxSpeakerID SpeakerID = 1

type Model = uint8

const (
	Model_UNICAST   Model = 0
	Model_MULTICAST Model = 1
)

type State uint32

const (
	State_OFFLINE    State = 0
	State_ONLINE     State = 0x00000001
	State_CONNERROR  State = 0x01000000
	State_QUEUEERROR State = 0x02000000
	State_DELETED    State = 0x80000000
)

type Statistic struct {
	Queue uint32 `jp:"q"` // 队列中数据量
	Spend uint64 `jp:"s"` // 已经发送的数据量
	Drop  uint32 `jp:"d"` // 被丢弃的数据量
	Error uint32 `jp:"e"`
}

type PowerState = uint8
