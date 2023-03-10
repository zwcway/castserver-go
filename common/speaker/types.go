package speaker

type LineID uint8
type ID uint32

var DefaultLineID LineID = 0

type Model uint8

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

func (id *ID) IsValid() bool {
	return true
}

type Statistic struct {
	Queue uint32 // 队列中数据量
	Spend uint64 // 已经发送的数据量
	Drop  uint32 // 被丢弃的数据量
	Error uint32
}

type PowerState uint8
