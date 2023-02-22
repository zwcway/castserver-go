package audio

import "errors"

type Channel uint8

const (
	AudioChannel_NONE Channel = iota
	AudioChannel_FRONT_LEFT
	AudioChannel_FRONT_RIGHT
	AudioChannel_FRONT_CENTER
	AudioChannel_FRONT_LEFT_OF_CENTER
	AudioChannel_FRONT_RIGHT_OF_CENTER
	AudioChannel_LOW_FREQUENCY
	AudioChannel_BACK_LEFT
	AudioChannel_BACK_RIGHT
	AudioChannel_BACK_CENTER
	AudioChannel_SIDE_LEFT
	AudioChannel_SIDE_RIGHT
	AudioChannel_TOP_CENTER
	AudioChannel_TOP_FRONT_LEFT
	AudioChannel_TOP_FRONT_CENTER
	AudioChannel_TOP_FRONT_RIGHT
	AudioChannel_TOP_BACK_LEFT
	AudioChannel_TOP_BACK_CENTER
	AudioChannel_TOP_BACK_RIGHT
	AudioChannel_MAX
)

func NewAudioChannel(i int32) Channel {
	var a Channel
	a.FromInt(i)
	return a
}

func (a *Channel) FromInt(i int32) {
	*a = Channel(i)
}

func (a *Channel) toInt() int32 {
	return int32(*a)
}

func (a *Channel) Name() string {
	switch *a {
	case AudioChannel_FRONT_LEFT:
		return "Front Left"
	case AudioChannel_FRONT_RIGHT:
		return "Front Right"
	case AudioChannel_FRONT_CENTER:
		return "Front Center"
	case AudioChannel_LOW_FREQUENCY:
		return "Subwoofer"
	case AudioChannel_BACK_LEFT:
		return "Rear Left"
	case AudioChannel_BACK_RIGHT:
		return "Rear Right"
	case AudioChannel_FRONT_LEFT_OF_CENTER:
		return "Front Left Of Center"
	case AudioChannel_FRONT_RIGHT_OF_CENTER:
		return "Front Right Of Center"
	case AudioChannel_BACK_CENTER:
		return "Rear Center"
	case AudioChannel_SIDE_LEFT:
		return "Side Left"
	case AudioChannel_SIDE_RIGHT:
		return "Side Right"
	case AudioChannel_TOP_CENTER:
		return "Top Center"
	case AudioChannel_TOP_FRONT_LEFT:
		return "Top Front Left"
	case AudioChannel_TOP_FRONT_CENTER:
		return "Top Front Center"
	case AudioChannel_TOP_FRONT_RIGHT:
		return "Top Front Right"
	case AudioChannel_TOP_BACK_LEFT:
		return "Top Rear Left"
	case AudioChannel_TOP_BACK_CENTER:
		return "Top Rear Center"
	case AudioChannel_TOP_BACK_RIGHT:
		return "Top Rear Right"
	default:
		return "Unknown"
	}
}

func (a *Channel) IsValid() bool {
	return *a > AudioChannel_NONE && *a < AudioChannel_MAX
}

type ChannelMask uint32

func NewAudioChannelMask(arr []uint8) (ChannelMask, error) {
	var a ChannelMask
	err := a.FromSlice(arr)
	return a, err
}

func (m *ChannelMask) FromSlice(arr []uint8) error {
	if len(arr) > 32 {
		return errors.New("channels too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func (m *ChannelMask) FromChannelSlice(arr []Channel) error {
	if len(arr) > 32 {
		return errors.New("channels too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func (m *ChannelMask) Count() int {
	c := 0
	for i := 0; i < 32; i++ {
		if (*m>>i)&0x01 > 0 {
			c++
		}
	}
	return c
}

func (m *ChannelMask) Isset(a uint8) bool {
	return maskIsset(uint(*m), a)
}

func (m *ChannelMask) IssetSlice(a []uint8) bool {
	return maskIssetSlice(uint(*m), a)
}

func (m *ChannelMask) CombineSlice(a []uint8) bool {
	r := maskCombineSlice(uint(*m), a)
	*m = ChannelMask(r)
	return r > 0
}

func (m *ChannelMask) IsValid() bool {
	return *m > 0 && ((*m)>>(AudioChannel_MAX-1)) == 0
}

func (m *ChannelMask) Slice() []int32 {
	s := []int32{}
	for i := 0; i < 16; i++ {
		if (*m>>i)&0x01 == 1 {
			b := Channel(i + 1)
			s = append(s, b.toInt())
		}
	}
	return s
}

var ChannelLayout20 ChannelLayout = NewChannelLayout([]Channel{AudioChannel_FRONT_LEFT, AudioChannel_FRONT_RIGHT})

func NewChannelLayout(ch []Channel) ChannelLayout {
	var a ChannelLayout
	a.Mask.FromChannelSlice(ch)
	a.Count = a.Mask.Count()
	return a
}

type ChannelLayout struct {
	Mask  ChannelMask
	Count int
}

func (l *ChannelLayout) String() string {
	return ""
}
