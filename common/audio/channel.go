package audio

import "errors"

type Channel uint8

const (
	ChannelNone Channel = iota
	ChannelFrontLeft
	ChannelFrontRight
	ChannelFrontCenter
	ChannelFrontLeftOfCenter
	ChannelFrontRightOfCenter
	ChannelLowFrequency
	ChannelBackLeft
	ChannelBackRight
	ChannelBackCenter
	ChannelSideLeft
	ChannelSideRight
	ChannelTopFrontLeft
	ChannelTopFrontCenter
	ChannelTopFrontRight
	ChannelTopBackLeft
	ChannelTopBackCenter
	ChannelTopBackRight
	ChannelMax
)

func NewAudioChannel(i int) Channel {
	return Channel(i)
}

func ChannelFromInt(i int) Channel {
	return Channel(i)
}

func (a Channel) ToInt() int {
	return int(a)
}

func (a Channel) String() string {
	switch a {
	case ChannelFrontLeft:
		return "Front Left"
	case ChannelFrontRight:
		return "Front Right"
	case ChannelFrontCenter:
		return "Front Center"
	case ChannelLowFrequency:
		return "Subwoofer"
	case ChannelBackLeft:
		return "Rear Left"
	case ChannelBackRight:
		return "Rear Right"
	case ChannelFrontLeftOfCenter:
		return "Front Left Of Center"
	case ChannelFrontRightOfCenter:
		return "Front Right Of Center"
	case ChannelBackCenter:
		return "Rear Center"
	case ChannelSideLeft:
		return "Side Left"
	case ChannelSideRight:
		return "Side Right"
	case ChannelTopFrontLeft:
		return "Top Front Left"
	case ChannelTopFrontCenter:
		return "Top Front Center"
	case ChannelTopFrontRight:
		return "Top Front Right"
	case ChannelTopBackLeft:
		return "Top Rear Left"
	case ChannelTopBackCenter:
		return "Top Rear Center"
	case ChannelTopBackRight:
		return "Top Rear Right"
	default:
		return "Unknown"
	}
}

func (a Channel) IsValid() bool {
	return a > ChannelNone && a < ChannelMax
}

type ChannelIndex = []int8
type ChannelMask uint32

func NewChannelMask(arr []uint8) (ChannelMask, error) {
	return ChannelMaskFromSlice(arr)
}

func ChannelMaskFromSlice(arr []uint8) (ChannelMask, error) {
	if len(arr) > 32 {
		return 0, errors.New("channels too large")
	}
	m := ChannelMask(0)
	for _, a := range arr {
		m = ChannelMaskAdd(m, Channel(a))
	}
	return m, nil
}

func ChannelMaskAdd(m ChannelMask, ch Channel) ChannelMask {
	if ch.IsValid() {
		m |= 1 << (ch - 1)
	}
	return m
}

func ChannelMaskFromChannelSlice(arr []Channel) (ChannelMask, error) {
	if len(arr) > int(ChannelMax) {
		return 0, errors.New("channels too large")
	}
	m := ChannelMask(0)
	for _, a := range arr {
		m = ChannelMaskAdd(m, a)
	}
	return m, nil
}

func (m ChannelMask) Count() int {
	c := 0
	for i := 0; i < int(ChannelMax); i++ {
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

func (m ChannelMask) IntersectIntSlice(a []uint8) bool {
	r, err := NewChannelMask(a)
	if err != nil {
		return false
	}
	m &= r
	return m.IsValid()
}

// IntersectSlice 默认参数合法
func (m ChannelMask) IntersectSlice(a []Channel) ChannelMask {
	m &= ChannelMask(MakeMaskFromSlice(a))
	return m
}

func (m ChannelMask) Intersect(a ChannelMask) ChannelMask {
	m &= a
	return m
}

// CombineSlice 默认参数合法
func (m ChannelMask) CombineSlice(a []Channel) ChannelMask {
	m &= ChannelMask(MakeMaskFromSlice(a))
	return m
}

func (m ChannelMask) Combine(a ChannelMask) ChannelMask {
	m &= a
	return m
}

func (m ChannelMask) IsValid() bool {
	return m > 0 && ((m)>>(ChannelMax-1)) == 0
}

func (m ChannelMask) Slice() []Channel {
	s := make([]Channel, ChannelMax)
	j := 0
	for i := 0; i < int(ChannelMax); i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = Channel(i + 1)
			j++
		}
	}
	return s[:j]
}

func (m ChannelMask) SliceInt() []int {
	s := make([]int, ChannelMax)
	j := 0
	for i := 0; i < int(ChannelMax); i++ {
		if (m>>i)&0x01 == 1 {
			s[j] = i + 1
			j++
		}
	}
	return s[:j]
}

func (m ChannelMask) ChannelIndex() ChannelIndex {
	ci := make(ChannelIndex, ChannelMax)
	for i := 0; i < int(ChannelMax); i++ {
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
	Layout10 = NewLayout(ChannelFrontCenter) // 前中
	Layout20 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight) // 前左、前右
	Layout21 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelLowFrequency) // 前左、前右，低音
	Layout22 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelSideLeft,
		ChannelSideRight) // 前左、前右，环左、环右
	Layout30 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter) // 前左、前右、前中
	Layout31 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelLowFrequency) // 前左、前右、前中，低音
	Layout40 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelBackCenter) // 前左、前右、前中、后中
	Layout41 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelBackCenter,
		ChannelLowFrequency) // 前左、前右、前中、后中，低音
	Layout50 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight) // 前左、前右、前中、环左、环右
	Layout51 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelLowFrequency) // 前左、前右、前中、环左、环右，低音
	Layout5B0 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelBackLeft,
		ChannelBackRight) // 前左、前右、前中、后左、后右
	Layout5B1 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelBackLeft,
		ChannelBackRight,
		ChannelLowFrequency) // 前左、前右、前中、后左、后右，低音
	Layout60 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelBackCenter) // 前左、前右、前中、环左、环右、后中
	Layout61 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelLowFrequency,
		ChannelBackCenter) // 前左、前右、前中、环左、环右、后中，低音
	Layout70 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelBackLeft,
		ChannelBackRight) // 前左、前右、前中、环左、环右、后左、后右
	Layout71 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelBackLeft,
		ChannelBackRight,
		ChannelLowFrequency) // 前左、前右、前中、环左、环右、后左、后右，低音
	Layout702 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelBackLeft,
		ChannelBackRight,
		ChannelTopFrontLeft,
		ChannelTopFrontRight) // 前左、前右、前中、环左、环右、后左、后右，上前左、上前右
	Layout712 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelBackLeft,
		ChannelBackRight,
		ChannelLowFrequency,
		ChannelTopFrontLeft,
		ChannelTopFrontRight) // 前左、前右、前中、环左、环右、后左、后右，低音，上前左、上前右
	Layout714 = NewLayout(
		ChannelFrontLeft,
		ChannelFrontRight,
		ChannelFrontCenter,
		ChannelSideLeft,
		ChannelSideRight,
		ChannelBackLeft,
		ChannelBackRight,
		ChannelLowFrequency,
		ChannelTopFrontLeft,
		ChannelTopFrontRight,
		ChannelTopBackLeft,
		ChannelTopBackRight) // 前左、前右、前中、环左、环右、后左、后右，低音，上前左、上前右、上后左、上后右

	LayoutMono   = Layout10
	LayoutStereo = Layout20

	ChannalLayoutMAX = newMaxLayout()
)

func newMaxLayout() Layout {
	var a Layout
	for ch := ChannelNone + 1; ch < ChannelMax; ch++ {
		a.ChannelMask = ChannelMaskAdd(a.ChannelMask, ch)
	}
	a.Count = uint32(a.ChannelMask.Count())
	return a
}

func NewLayout(ch ...Channel) Layout {
	var a Layout
	a.ChannelMask, _ = ChannelMaskFromChannelSlice(ch)
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
	a.ChannelMask, _ = ChannelMaskFromChannelSlice(append(a.ChannelMask.Slice(), ch...))
	a.Count = uint32(a.ChannelMask.Count())
	return a
}

type ChannelRoute struct {
	From []Channel
	To   Channel
}
