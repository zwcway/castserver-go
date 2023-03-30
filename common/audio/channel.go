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
	s := []Channel{}
	for i := 0; i < 32; i++ {
		if (m>>i)&0x01 == 1 {
			b := Channel(i + 1)
			s = append(s, b)
		}
	}
	return s
}

func (m ChannelMask) SliceInt() []int {
	s := m.Slice()
	si := make([]int, len(s))
	for i, ss := range s {
		si[i] = ss.ToInt()
	}
	return si
}

var (
	ChannelLayout10  = NewChannelLayout(Channel_FRONT_CENTER)                                         // 前中
	ChannelLayout20  = NewChannelLayout(Channel_FRONT_LEFT, Channel_FRONT_RIGHT)                      // 前左、前右
	ChannelLayout21  = extendLayout(ChannelLayout20, Channel_LOW_FREQUENCY)                           // 前左、前右，低音
	ChannelLayout22  = extendLayout(ChannelLayout20, Channel_SIDE_LEFT, Channel_SIDE_RIGHT)           // 前左、前右，环左、环右
	ChannelLayout30  = extendLayout(ChannelLayout20, Channel_FRONT_CENTER)                            // 前左、前右、前中
	ChannelLayout31  = extendLayout(ChannelLayout30, Channel_LOW_FREQUENCY)                           // 前左、前右、前中，低音
	ChannelLayout40  = extendLayout(ChannelLayout30, Channel_BACK_CENTER)                             // 前左、前右、前中、后中
	ChannelLayout41  = extendLayout(ChannelLayout40, Channel_LOW_FREQUENCY)                           // 前左、前右、前中、后中，低音
	ChannelLayout50  = extendLayout(ChannelLayout30, Channel_SIDE_LEFT, Channel_SIDE_RIGHT)           // 前左、前右、前中、环左、环右
	ChannelLayout51  = extendLayout(ChannelLayout50, Channel_LOW_FREQUENCY)                           // 前左、前右、前中、环左、环右，低音
	ChannelLayout5B0 = extendLayout(ChannelLayout30, Channel_BACK_LEFT, Channel_BACK_RIGHT)           // 前左、前右、前中、后左、后右
	ChannelLayout5B1 = extendLayout(ChannelLayout5B0, Channel_LOW_FREQUENCY)                          // 前左、前右、前中、后左、后右，低音
	ChannelLayout60  = extendLayout(ChannelLayout50, Channel_BACK_CENTER)                             // 前左、前右、前中、环左、环右、后中
	ChannelLayout61  = extendLayout(ChannelLayout51, Channel_BACK_CENTER)                             // 前左、前右、前中、环左、环右、后中，低音
	ChannelLayout70  = extendLayout(ChannelLayout50, Channel_BACK_LEFT, Channel_BACK_RIGHT)           // 前左、前右、前中、环左、环右、后左、后右
	ChannelLayout71  = extendLayout(ChannelLayout70, Channel_LOW_FREQUENCY)                           // 前左、前右、前中、环左、环右、后左、后右，低音
	ChannelLayout702 = extendLayout(ChannelLayout70, Channel_TOP_FRONT_LEFT, Channel_TOP_FRONT_RIGHT) // 前左、前右、前中、环左、环右、后左、后右，上前左、上前右
	ChannelLayout712 = extendLayout(ChannelLayout71, Channel_TOP_FRONT_LEFT, Channel_TOP_FRONT_RIGHT) // 前左、前右、前中、环左、环右、后左、后右，低音，上前左、上前右
	ChannelLayout714 = extendLayout(ChannelLayout712, Channel_TOP_BACK_LEFT, Channel_TOP_BACK_RIGHT)  // 前左、前右、前中、环左、环右、后左、后右，低音，上前左、上前右、上后左、上后右

	ChannelLayoutMono   = ChannelLayout10
	ChannelLayoutStereo = ChannelLayout20

	ChannalLayoutMAX = newMaxLayout()
)

func newMaxLayout() ChannelLayout {
	var a ChannelLayout
	for ch := Channel_NONE + 1; ch < Channel_MAX; ch++ {
		a.Mask.Add(ch)
	}
	a.Count = a.Mask.Count()
	return a
}

func NewChannelLayout(ch ...Channel) ChannelLayout {
	var a ChannelLayout
	a.Mask.FromChannelSlice(ch)
	a.Count = a.Mask.Count()
	return a
}

type ChannelLayout struct {
	Mask  ChannelMask
	Count int
}

func (l ChannelLayout) String() string {
	switch l.Mask {
	case ChannelLayout10.Mask:
		return "mono"
	case ChannelLayout20.Mask:
		return "stereo"
	case ChannelLayout21.Mask:
		return "2.1"
	case ChannelLayout22.Mask:
		return "2.2"
	case ChannelLayout30.Mask:
		return "3.0"
	case ChannelLayout31.Mask:
		return "3.1"
	case ChannelLayout40.Mask:
		return "4.0"
	case ChannelLayout41.Mask:
		return "4.1"
	case ChannelLayout50.Mask:
		return "5.0"
	case ChannelLayout51.Mask:
		return "5.1"
	case ChannelLayout5B0.Mask:
		return "5.0(back)"
	case ChannelLayout5B1.Mask:
		return "5.1(back)"
	case ChannelLayout60.Mask:
		return "6.0"
	case ChannelLayout61.Mask:
		return "6.1"
	case ChannelLayout70.Mask:
		return "7.0"
	case ChannelLayout71.Mask:
		return "7.1"
	case ChannelLayout702.Mask:
		return "7.0.2"
	case ChannelLayout712.Mask:
		return "7.1.2"
	case ChannelLayout714.Mask:
		return "7.1.4"
	}
	return ""
}

func (l ChannelLayout) IsValid() bool {
	return l.Mask.IsValid() && l.Count == l.Mask.Count()
}

func (l ChannelLayout) Channels() []Channel {
	return l.Mask.Slice()
}

func (l ChannelLayout) Equal(r ChannelLayout) bool {
	return l.Mask == r.Mask
}

func extendLayout(a ChannelLayout, ch ...Channel) ChannelLayout {
	a.Mask.FromChannelSlice(append(a.Mask.Slice(), ch...))
	a.Count = a.Mask.Count()
	return a
}

type ChannelRoute struct {
	From []Channel
	To   Channel
}
