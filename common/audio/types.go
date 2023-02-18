package audio

import (
	"errors"
	"fmt"
)

type AudioEnum interface {
	FromInt(int32) error
	ToInt() int32
	Name() string
	IsValid() bool
}

type AudioBits uint8

const (
	AudioBits_NONE AudioBits = iota
	AudioBits_U8
	AudioBits_S16LE
	AudioBits_U16LE
	AudioBits_S24LE
	AudioBits_U24LE
	AudioBits_S32LE
	AudioBits_U32LE
	AudioBits_32LEF
	AudioBits_MAX
)

func NewAudioBits(i int32) (AudioBits, error) {
	var a AudioBits
	err := a.FromInt(i)
	return a, err
}
func (a *AudioBits) FromInt(i int32) error {
	switch i {
	case 8:
		*a = AudioBits_U8
	case 16:
		*a = AudioBits_U16LE
	case 24:
		*a = AudioBits_U24LE
	case 32:
		*a = AudioBits_U32LE
	default:
		*a = AudioBits_NONE
		return fmt.Errorf("unknown Bits")
	}
	return nil
}

func (AudioBits) FromIntSlice(i []int32) ([]AudioBits, error) {
	var (
		s   []AudioBits
		a   AudioBits
		err error
	)
	for _, v := range i {
		err = a.FromInt(v)
		if err != nil {
			return nil, err
		}
		s = append(s, a)
	}

	return s, nil
}

func (a *AudioBits) FromBytes(i int8) error {
	return a.FromInt(int32(i * 8))
}

func (a *AudioBits) ToInt() int32 {
	switch *a {
	case AudioBits_U8:
		return 8
	case AudioBits_U16LE:
		return 16
	case AudioBits_U24LE:
		return 24
	case AudioBits_U32LE:
		return 32
	case AudioBits_32LEF:
		return 32
	default:
		return 0
	}
}

func (a *AudioBits) FromName(n string) {
	switch n {
	case "u8":
		*a = AudioBits_U8
	case "s16le":
		*a = AudioBits_S16LE
	case "u16le":
		*a = AudioBits_U16LE
	case "s24le":
		*a = AudioBits_S24LE
	case "u24le":
		*a = AudioBits_U24LE
	case "s32le":
		*a = AudioBits_S32LE
	case "u32le":
		*a = AudioBits_U32LE
	case "fltle":
		*a = AudioBits_32LEF
	default:
		*a = AudioBits_NONE
	}
}

func (a *AudioBits) Name() string {
	switch *a {
	case AudioBits_U8:
		return "u8"
	case AudioBits_S16LE:
		return "s16le"
	case AudioBits_U16LE:
		return "u16le"
	case AudioBits_S24LE:
		return "s24le"
	case AudioBits_U24LE:
		return "u24le"
	case AudioBits_S32LE:
		return "s32le"
	case AudioBits_U32LE:
		return "u32le"
	case AudioBits_32LEF:
		return "fltle"
	default:
		return "Unknown"
	}
}

func (a *AudioBits) Bits() int32 {
	return a.ToInt()
}
func (a *AudioBits) Size() int32 {
	return int32(a.ToInt() / 8)
}

type AudioRate uint8

const (
	AudioRate_NONE AudioRate = iota
	AudioRate_44100
	AudioRate_48000
	AudioRate_96000
	AudioRate_192000
	AudioRate_384000
	AudioRate_MAX
)

func NewAudioRate(i int32) (AudioRate, error) {
	var a AudioRate
	err := a.FromInt(i)
	return a, err
}

func (a *AudioRate) FromInt(i int32) error {
	switch i {
	case 44100:
		*a = AudioRate_44100
	case 48000:
		*a = AudioRate_48000
	case 96000:
		*a = AudioRate_96000
	case 192000:
		*a = AudioRate_192000
	case 384000:
		*a = AudioRate_384000
	default:
		*a = AudioRate_NONE
		return errors.New("unknown Rate")
	}
	return nil
}

func (a *AudioRate) ToInt() int32 {
	switch *a {
	case AudioRate_44100:
		return 44100
	case AudioRate_48000:
		return 48000
	case AudioRate_96000:
		return 96000
	case AudioRate_192000:
		return 192000
	case AudioRate_384000:
		return 384000
	default:
		return 0
	}
}

type AudioChannel uint8

const (
	AudioChannel_NONE AudioChannel = iota
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

func NewAudioChannel(i int32) (AudioChannel, error) {
	var a AudioChannel
	err := a.FromInt(i)
	return a, err
}

func (a *AudioChannel) FromInt(i int32) error {
	*a = AudioChannel(i)
	return nil
}

func (a *AudioChannel) toInt() int32 {
	return int32(*a)
}

func (a *AudioChannel) Name() string {
	switch *a {
	case AudioChannel_FRONT_LEFT:
		return "Front Left"
	case AudioChannel_FRONT_RIGHT:
		return "Front Right"
	case AudioChannel_FRONT_CENTER:
		return "Front Center"
	case AudioChannel_LOW_FREQUENCY:
		return "LFE / Subwoofer"
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

func (a *AudioRate) IsValid() bool {
	return *a > AudioRate_NONE && *a < AudioRate_MAX
}

func (a *AudioBits) IsValid() bool {
	return *a > AudioBits_NONE && *a < AudioBits_MAX
}

func (a *AudioChannel) IsValid() bool {
	return *a > AudioChannel_NONE && *a < AudioChannel_MAX
}

type AudioMask interface {
	FromSlice([]uint8) error
	Isset(uint8) bool
	IsValid() bool
}
type AudioRateMask uint16
type AudioBitsMask uint16
type AudioChannelMask uint32

func NewAudioRateMask(arr []uint8) (*AudioRateMask, error) {
	var a AudioRateMask
	err := a.FromSlice(arr)
	return &a, err
}

func NewAudioBitsMask(arr []uint8) (*AudioBitsMask, error) {
	var a AudioBitsMask
	err := a.FromSlice(arr)
	return &a, err
}

func NewAudioChannelMask(arr []uint8) (*AudioChannelMask, error) {
	var a AudioChannelMask
	err := a.FromSlice(arr)
	return &a, err
}

func (m *AudioRateMask) FromSlice(arr []uint8) error {
	if len(arr) > 16 {
		return errors.New("Rates too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func (m *AudioBitsMask) FromSlice(arr []uint8) error {
	if len(arr) > 16 {
		return errors.New("Bits too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func (m *AudioChannelMask) FromSlice(arr []uint8) error {
	if len(arr) > 32 {
		return errors.New("Channels too large")
	}
	for _, a := range arr {
		*m |= 1 << (a - 1)
	}
	return nil
}

func maskIsset(m uint, a uint8) bool {
	return ((m) & (1 << a)) > 0
}

func (m *AudioRateMask) Isset(a uint8) bool {
	return maskIsset(uint(*m), a)
}

func (m *AudioBitsMask) Isset(a uint8) bool {
	return maskIsset(uint(*m), a)
}

func (m *AudioChannelMask) Isset(a uint8) bool {
	return maskIsset(uint(*m), a)
}

func maskIssetSlice(m uint, a []uint8) bool {
	for _, v := range a {
		if !maskIsset(m, uint8(v)) {
			return false
		}
	}

	return true
}

func (m *AudioRateMask) IssetSlice(a []uint8) bool {
	return maskIssetSlice(uint(*m), a)
}

func (m *AudioBitsMask) IssetSlice(a []uint8) bool {
	return maskIssetSlice(uint(*m), a)
}

func (m *AudioChannelMask) IssetSlice(a []uint8) bool {
	return maskIssetSlice(uint(*m), a)
}

func (m *AudioRateMask) IsValid() bool {
	return *m > 0 && ((*m)>>(AudioRate_MAX-1)) == 0
}

func (m *AudioBitsMask) IsValid() bool {
	return *m > 0 && ((*m)>>(AudioBits_MAX-1)) == 0
}

func (m *AudioChannelMask) IsValid() bool {
	return *m > 0 && ((*m)>>(AudioChannel_MAX-1)) == 0
}

type AudioFormat struct {
	channels      AudioChannelMask
	samplesPerSec uint32
	bitsPerSample uint32
}
