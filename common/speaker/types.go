package speaker

type LineID uint8

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

type ID uint32

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

type speakerMapSlice map[int][]*Speaker

func (s *speakerMapSlice) remove(key int, sp *Speaker) {
	slice, ok := (*s)[key]
	if !ok {
		return
	}
	for i, item := range slice {
		if item == sp {
			(*s)[key] = append(slice[:i], slice[i+1:]...)
			return
		}
	}
}

func (s *speakerMapSlice) len(key int) int {
	return len((*s)[key])
}

func (s *speakerMapSlice) add(key int, sp *Speaker) {
	if _, ok := (*s)[key]; ok {
		(*s)[key] = append((*s)[key], sp)
	} else {
		(*s)[key] = append([]*Speaker{}, sp)
	}
}
