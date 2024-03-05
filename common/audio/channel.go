package audio

import "errors"

type Channel uint8

const (
	Channel_NONE Channel = iota
	Channel_FRONT_LEFT
	Channel_FRONT_RIGHT
	Channel_FRONT_CENTER
	Channel_FRONT_LEFT_OF_CENTER
	Channel_FRONT_RIGHT_OF_CENTER
	Channel_LOW_FREQUENCY
	Channel_BACK_LEFT
	Channel_BACK_RIGHT
	Channel_BACK_CENTER
	Channel_SIDE_LEFT
	Channel_SIDE_RIGHT
	Channel_TOP_FRONT_LEFT
	Channel_TOP_FRONT_CENTER
	Channel_TOP_FRONT_RIGHT
	Channel_TOP_BACK_LEFT
	Channel_TOP_BACK_CENTER
	Channel_TOP_BACK_RIGHT
	Channel_MAX
)

func NewAudioChannel(i int) Channel {
	var a Channel
	a.FromInt(i)
	return a
}

func (a *Channel) FromInt(i int) {
	*a = Channel(i)
}

func (a Channel) ToInt() int {
	return int(a)
}

func (a Channel) String() string {
	switch a {
	case Channel_FRONT_LEFT:
		return "Front Left"
	case Channel_FRONT_RIGHT:
		return "Front Right"
	case Channel_FRONT_CENTER:
		return "Front Center"
	case Channel_LOW_FREQUENCY:
		return "Subwoofer"
	case Channel_BACK_LEFT:
		return "Rear Left"
	case Channel_BACK_RIGHT:
		return "Rear Right"
	case Channel_FRONT_LEFT_OF_CENTER:
		return "Front Left Of Center"
	case Channel_FRONT_RIGHT_OF_CENTER:
		return "Front Right Of Center"
	case Channel_BACK_CENTER:
		return "Rear Center"
	case Channel_SIDE_LEFT:
		return "Side Left"
	case Channel_SIDE_RIGHT:
		return "Side Right"
	case Channel_TOP_FRONT_LEFT:
		return "Top Front Left"
	case Channel_TOP_FRONT_CENTER:
		return "Top Front Center"
	case Channel_TOP_FRONT_RIGHT:
		return "Top Front Right"
	case Channel_TOP_BACK_LEFT:
		return "Top Rear Left"
	case Channel_TOP_BACK_CENTER:
		return "Top Rear Center"
	case Channel_TOP_BACK_RIGHT:
		return "Top Rear Right"
	default:
		return "Unknown"
	}
}

func (a Channel) IsValid() bool {
	return a > Channel_NONE && a < Channel_MAX
}

type ChannelIndex = []int8
type ChannelMask uint32

func NewChannelMask(arr []uint8) (ChannelMask, error) {
	var a ChannelMask
	err := a.FromSlice(arr)
	return a, err
}

func (m *ChannelMask) FromSlice(arr []uint8) error {
	if len(arr) > 32 {
		return errors.New("channels too large")
	}
	*m = 0
	for _, a := range arr {
		m.Add(Channel(a))
	}
	return nil
}

func (m *ChannelMask) Add(ch Channel) {
	if ch.IsValid() {
		*m |= 1 << (ch - 1)
	}
}

func (m *ChannelMask) FromChannelSlice(arr []Channel) error {
	if len(arr) > int(Channel_MAX) {
		return errors.New("channels too large")
	}
	*m = 0
	for _, a := range arr {
		m.Add(a)
	}
	return nil
}

func (m ChannelMask) Count() int {
	c := 0
	for i := 0; i < int(Channel_MAX); i++ {
		if (m>>i)&0x01 > 0 {
			c++
		}
	}
	return c
}

func (m ChannelMask) IssetInt(a uint8) bool {
	return MaskIsset(uint32(m), a)
}

func (m ChannelMask) Isset(a Channel) bool {
	return MaskIsset(uint32(m), uint8(a))
}

func (m ChannelMask) IssetSlice(a []uint8) bool {
	return MaskIssetIntSlice(uint32(m), a)
}

func (m *ChannelMask) IntersectIntSlice(a []uint8) bool {
	r, err := NewChannelMask(a)
	if err != nil {
		return false
	}
	*m &= r
	return m.IsValid()
}

// 默认参数合法
func (m *ChannelMask) IntersectSlice(a []Channel) bool {
	*m &= ChannelMask(MakeMaskFromSlice(a))
	return m.IsValid()
}

func (m *ChannelMask) Intersect(a ChannelMask) bool {
	*m &= a
	return m.IsValid()
}

// 默认参数合法
func (m *ChannelMask) CombineSlice(a []Channel) bool {
	*m &= ChannelMask(MakeMaskFromSlice(a))
	return m.IsValid()
}

func (m *ChannelMask) Combine(a ChannelMask) bool {
	*m &= a
	return m.IsValid()
}

func (m ChannelMask) IsValid() bool {
	return m > 0 && ((m)>>(Channel_MAX-1)) == 0
}

func (m ChannelMask) Slice() []Channel {
	s := make([]Channel, Channel_MAX)
	j := 0
	for i := 0; i < int(Channel_MAX); i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = Channel(i + 1)
			j++
		}
	}
	return s[:j]
}

func (m ChannelMask) SliceInt() []int {
	s := make([]int, Channel_MAX)
	j := 0
	for i := 0; i < int(Channel_MAX); i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = i + 1
			j++
		}
	}
	return s[:j]
}

func (m ChannelMask) ChannelIndex() ChannelIndex {
	ci := make(ChannelIndex, Channel_MAX)
	for i := 0; i < int(Channel_MAX); i++ {
		ci[i] = -1
	}
	if m > 0 {
		for i, ch := range m.Slice() {
			ci[ch] = int8(i)
		}
	}
	return ci
}

var (
	Layout10 = NewLayout(Channel_FRONT_CENTER) // 前中
	Layout20 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT) // 前左、前右
	Layout21 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_LOW_FREQUENCY) // 前左、前右，低音
	Layout22 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT) // 前左、前右，环左、环右
	Layout30 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER) // 前左、前右、前中
	Layout31 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_LOW_FREQUENCY) // 前左、前右、前中，低音
	Layout40 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_BACK_CENTER) // 前左、前右、前中、后中
	Layout41 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_BACK_CENTER,
		Channel_LOW_FREQUENCY) // 前左、前右、前中、后中，低音
	Layout50 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT) // 前左、前右、前中、环左、环右
	Layout51 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_LOW_FREQUENCY) // 前左、前右、前中、环左、环右，低音
	Layout5B0 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_BACK_LEFT,
		Channel_BACK_RIGHT) // 前左、前右、前中、后左、后右
	Layout5B1 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_BACK_LEFT,
		Channel_BACK_RIGHT,
		Channel_LOW_FREQUENCY) // 前左、前右、前中、后左、后右，低音
	Layout60 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_BACK_CENTER) // 前左、前右、前中、环左、环右、后中
	Layout61 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_LOW_FREQUENCY,
		Channel_BACK_CENTER) // 前左、前右、前中、环左、环右、后中，低音
	Layout70 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_BACK_LEFT,
		Channel_BACK_RIGHT) // 前左、前右、前中、环左、环右、后左、后右
	Layout71 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_BACK_LEFT,
		Channel_BACK_RIGHT,
		Channel_LOW_FREQUENCY) // 前左、前右、前中、环左、环右、后左、后右，低音
	Layout702 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_BACK_LEFT,
		Channel_BACK_RIGHT,
		Channel_TOP_FRONT_LEFT,
		Channel_TOP_FRONT_RIGHT) // 前左、前右、前中、环左、环右、后左、后右，上前左、上前右
	Layout712 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_BACK_LEFT,
		Channel_BACK_RIGHT,
		Channel_LOW_FREQUENCY,
		Channel_TOP_FRONT_LEFT,
		Channel_TOP_FRONT_RIGHT) // 前左、前右、前中、环左、环右、后左、后右，低音，上前左、上前右
	Layout714 = NewLayout(
		Channel_FRONT_LEFT,
		Channel_FRONT_RIGHT,
		Channel_FRONT_CENTER,
		Channel_SIDE_LEFT,
		Channel_SIDE_RIGHT,
		Channel_BACK_LEFT,
		Channel_BACK_RIGHT,
		Channel_LOW_FREQUENCY,
		Channel_TOP_FRONT_LEFT,
		Channel_TOP_FRONT_RIGHT,
		Channel_TOP_BACK_LEFT,
		Channel_TOP_BACK_RIGHT) // 前左、前右、前中、环左、环右、后左、后右，低音，上前左、上前右、上后左、上后右

	LayoutMono   = Layout10
	LayoutStereo = Layout20

	ChannalLayoutMAX = newMaxLayout()
)

func newMaxLayout() Layout {
	var a Layout
	for ch := Channel_NONE + 1; ch < Channel_MAX; ch++ {
		a.ChannelMask.Add(ch)
	}
	a.Count = uint32(a.ChannelMask.Count())
	return a
}

func NewLayout(ch ...Channel) Layout {
	var a Layout
	a.ChannelMask.FromChannelSlice(ch)
	a.Count = uint32(a.ChannelMask.Count())
	return a
}

type Layout struct {
	ChannelMask
	Count uint32
}

func (l Layout) String() string {
	switch l.ChannelMask {
	case Layout10.ChannelMask:
		return "mono"
	case Layout20.ChannelMask:
		return "stereo"
	case Layout21.ChannelMask:
		return "2.1"
	case Layout22.ChannelMask:
		return "2.2"
	case Layout30.ChannelMask:
		return "3.0"
	case Layout31.ChannelMask:
		return "3.1"
	case Layout40.ChannelMask:
		return "4.0"
	case Layout41.ChannelMask:
		return "4.1"
	case Layout50.ChannelMask:
		return "5.0"
	case Layout51.ChannelMask:
		return "5.1"
	case Layout5B0.ChannelMask:
		return "5.0(back)"
	case Layout5B1.ChannelMask:
		return "5.1(back)"
	case Layout60.ChannelMask:
		return "6.0"
	case Layout61.ChannelMask:
		return "6.1"
	case Layout70.ChannelMask:
		return "7.0"
	case Layout71.ChannelMask:
		return "7.1"
	case Layout702.ChannelMask:
		return "7.0.2"
	case Layout712.ChannelMask:
		return "7.1.2"
	case Layout714.ChannelMask:
		return "7.1.4"
	}
	return ""
}

func (l Layout) Channels() []Channel {
	return l.ChannelMask.Slice()
}

func (l Layout) Equal(r Layout) bool {
	return l.ChannelMask == r.ChannelMask
}

func extendLayout(a Layout, ch ...Channel) Layout {
	a.ChannelMask.FromChannelSlice(append(a.ChannelMask.Slice(), ch...))
	a.Count = uint32(a.ChannelMask.Count())
	return a
}

type ChannelRoute struct {
	From []Channel
	To   Channel
}
