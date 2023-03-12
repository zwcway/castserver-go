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
	Channel_TOP_CENTER
	Channel_TOP_FRONT_LEFT
	Channel_TOP_FRONT_CENTER
	Channel_TOP_FRONT_RIGHT
	Channel_TOP_BACK_LEFT
	Channel_TOP_BACK_CENTER
	Channel_TOP_BACK_RIGHT
	Channel_MAX
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
	case Channel_TOP_CENTER:
		return "Top Center"
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

func (a *Channel) IsValid() bool {
	return *a > Channel_NONE && *a < Channel_MAX
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
	return *m > 0 && ((*m)>>(Channel_MAX-1)) == 0
}

func (m *ChannelMask) Slice() []Channel {
	s := []Channel{}
	for i := 0; i < 16; i++ {
		if (*m>>i)&0x01 == 1 {
			b := Channel(i + 1)
			s = append(s, b)
		}
	}
	return s
}

var (
	ChannelLayout10  = NewChannelLayout(Channel_FRONT_CENTER)
	ChannelLayout20  = NewChannelLayout(Channel_FRONT_LEFT, Channel_FRONT_RIGHT)
	ChannelLayout21  = extendLayout(ChannelLayout20, Channel_LOW_FREQUENCY)
	ChannelLayout22  = extendLayout(ChannelLayout20, Channel_SIDE_LEFT, Channel_SIDE_RIGHT)
	ChannelLayout30  = extendLayout(ChannelLayout20, Channel_FRONT_CENTER)
	ChannelLayout31  = extendLayout(ChannelLayout30, Channel_LOW_FREQUENCY)
	ChannelLayout40  = extendLayout(ChannelLayout30, Channel_BACK_CENTER)
	ChannelLayout41  = extendLayout(ChannelLayout40, Channel_LOW_FREQUENCY)
	ChannelLayout50  = extendLayout(ChannelLayout30, Channel_SIDE_LEFT, Channel_SIDE_RIGHT)
	ChannelLayout502 = extendLayout(ChannelLayout30, Channel_BACK_LEFT, Channel_BACK_RIGHT)
	ChannelLayout51  = extendLayout(ChannelLayout50, Channel_LOW_FREQUENCY)
	ChannelLayout512 = extendLayout(ChannelLayout502, Channel_LOW_FREQUENCY)
	ChannelLayout60  = extendLayout(ChannelLayout50, Channel_BACK_CENTER)
	ChannelLayout602 = extendLayout(ChannelLayout22, Channel_FRONT_LEFT_OF_CENTER, Channel_FRONT_RIGHT_OF_CENTER)
	ChannelLayout611 = extendLayout(ChannelLayout51, Channel_BACK_CENTER)
	ChannelLayout70  = extendLayout(ChannelLayout50, Channel_BACK_LEFT, Channel_BACK_RIGHT)
	ChannelLayout71  = extendLayout(ChannelLayout70, Channel_LOW_FREQUENCY)

	ChannelLayoutMONO   = ChannelLayout10
	ChannelLayoutSTEREO = ChannelLayout20
)

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

func (l *ChannelLayout) String() string {
	switch l.Mask {
	case ChannelLayout10.Mask:
		return "mono"
	case ChannelLayout20.Mask:
		return "stereo"
	case ChannelLayout21.Mask:
		return "2.1"
	case ChannelLayout30.Mask:
		return "3.0"
	case ChannelLayout22.Mask:
		return "2.2"
	case ChannelLayout50.Mask:
		return "5.0"
	case ChannelLayout51.Mask:
		return "5.1"
	case ChannelLayout70.Mask:
		return "7.0"
	case ChannelLayout71.Mask:
		return "7.1"
	}
	return ""
}

func extendLayout(a ChannelLayout, ch ...Channel) ChannelLayout {
	a.Mask.FromChannelSlice(append(a.Mask.Slice(), ch...))
	a.Count = a.Mask.Count()
	return a
}
