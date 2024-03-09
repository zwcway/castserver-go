package avutil

/*
#cgo pkg-config: libavutil libavformat

#include "avutil.c"
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
)

func NewErrorFromCCode(code int) error {
	size := C.size_t(256)
	buf := (*C.char)(C.av_mallocz(size))
	defer C.av_free(unsafe.Pointer(buf))

	if C.go_averror_is_eof(C.int(code)) == 1 {
		return &EofError{}
	}

	var err string
	if C.av_strerror(C.int(code), buf, size-1) == 0 {
		err = C.GoString(buf)
	} else {
		err = "Unknown error"
	}
	return &Error{
		code: (code),
		err:  err,
	}
}

type Error struct {
	code int
	err  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.code, e.err)
}

type EofError struct{}

func (e *EofError) Error() string {
	return "eof error"
}

type AllocError struct {
	Size int
}

func (e *AllocError) Error() string {
	return fmt.Sprintf("alloc[%d] error", e.Size)
}

func IsEof(err error) bool {
	_, ok := err.(*EofError)
	return ok
}

func channelFromAV(ch C.uint64_t) audio.Channel {
	switch ch {
	case C.AV_CH_FRONT_LEFT:
		return audio.Channel_FRONT_LEFT
	case C.AV_CH_FRONT_RIGHT:
		return audio.Channel_FRONT_RIGHT
	case C.AV_CH_FRONT_CENTER:
		return audio.Channel_FRONT_CENTER
	case C.AV_CH_LOW_FREQUENCY:
		return audio.Channel_LOW_FREQUENCY
	case C.AV_CH_BACK_LEFT:
		return audio.Channel_BACK_LEFT
	case C.AV_CH_BACK_RIGHT:
		return audio.Channel_BACK_RIGHT
	case C.AV_CH_FRONT_LEFT_OF_CENTER:
		return audio.Channel_FRONT_LEFT_OF_CENTER
	case C.AV_CH_FRONT_RIGHT_OF_CENTER:
		return audio.Channel_FRONT_RIGHT_OF_CENTER
	case C.AV_CH_BACK_CENTER:
		return audio.Channel_BACK_CENTER
	case C.AV_CH_SIDE_LEFT:
		return audio.Channel_SIDE_LEFT
	case C.AV_CH_SIDE_RIGHT:
		return audio.Channel_SIDE_RIGHT
	// case C.AV_CH_TOP_CENTER:
	// 	return audio.Channel_TOP_CENTER
	case C.AV_CH_TOP_FRONT_LEFT:
		return audio.Channel_TOP_FRONT_LEFT
	case C.AV_CH_TOP_FRONT_CENTER:
		return audio.Channel_TOP_FRONT_CENTER
	case C.AV_CH_TOP_FRONT_RIGHT:
		return audio.Channel_TOP_FRONT_RIGHT
	case C.AV_CH_TOP_BACK_LEFT:
		return audio.Channel_TOP_BACK_LEFT
	case C.AV_CH_TOP_BACK_CENTER:
		return audio.Channel_TOP_BACK_CENTER
	case C.AV_CH_TOP_BACK_RIGHT:
		return audio.Channel_TOP_BACK_RIGHT
	case C.AV_CH_STEREO_LEFT:
	case C.AV_CH_STEREO_RIGHT:
	case C.AV_CH_WIDE_LEFT:
	case C.AV_CH_WIDE_RIGHT:
	}
	return audio.Channel_NONE
}

func avFromChannel(ch audio.Channel) C.uint64_t {
	switch ch {
	case audio.Channel_FRONT_LEFT:
		return C.AV_CH_FRONT_LEFT
	case audio.Channel_FRONT_RIGHT:
		return C.AV_CH_FRONT_RIGHT
	case audio.Channel_FRONT_CENTER:
		return C.AV_CH_FRONT_CENTER
	case audio.Channel_LOW_FREQUENCY:
		return C.AV_CH_LOW_FREQUENCY
	case audio.Channel_BACK_LEFT:
		return C.AV_CH_BACK_LEFT
	case audio.Channel_BACK_RIGHT:
		return C.AV_CH_BACK_RIGHT
	case audio.Channel_FRONT_LEFT_OF_CENTER:
		return C.AV_CH_FRONT_LEFT_OF_CENTER
	case audio.Channel_FRONT_RIGHT_OF_CENTER:
		return C.AV_CH_FRONT_RIGHT_OF_CENTER
	case audio.Channel_BACK_CENTER:
		return C.AV_CH_BACK_CENTER
	case audio.Channel_SIDE_LEFT:
		return C.AV_CH_SIDE_LEFT
	case audio.Channel_SIDE_RIGHT:
		return C.AV_CH_SIDE_RIGHT
	// case audio.Channel_TOP_CENTER:
	// 	return C.AV_CH_TOP_CENTER
	case audio.Channel_TOP_FRONT_LEFT:
		return C.AV_CH_TOP_FRONT_LEFT
	case audio.Channel_TOP_FRONT_CENTER:
		return C.AV_CH_TOP_FRONT_CENTER
	case audio.Channel_TOP_FRONT_RIGHT:
		return C.AV_CH_TOP_FRONT_RIGHT
	case audio.Channel_TOP_BACK_LEFT:
		return C.AV_CH_TOP_BACK_LEFT
	case audio.Channel_TOP_BACK_CENTER:
		return C.AV_CH_TOP_BACK_CENTER
	case audio.Channel_TOP_BACK_RIGHT:
		return C.AV_CH_TOP_BACK_RIGHT
	}
	return 0
}

func BitsFromAVFormat(b C.enum_AVSampleFormat) audio.Bits {
	switch b {
	case C.AV_SAMPLE_FMT_U8, C.AV_SAMPLE_FMT_U8P:
		return audio.Bits_U8
	case C.AV_SAMPLE_FMT_S16, C.AV_SAMPLE_FMT_S16P:
		return audio.Bits_S16LE
	case C.AV_SAMPLE_FMT_S32, C.AV_SAMPLE_FMT_S32P:
		return audio.Bits_S32LE
	case C.AV_SAMPLE_FMT_FLT, C.AV_SAMPLE_FMT_FLTP:
		return audio.Bits_32LEF
	case C.AV_SAMPLE_FMT_DBL:
	case C.AV_SAMPLE_FMT_DBLP:
	case C.AV_SAMPLE_FMT_S64:
	case C.AV_SAMPLE_FMT_S64P:
	}
	return audio.Bits_NONE
}

func AVFormatFromBits(b audio.Bits) C.enum_AVSampleFormat {
	switch b {
	case audio.Bits_U8:
		return C.AV_SAMPLE_FMT_U8P
	case audio.Bits_S16LE:
		return C.AV_SAMPLE_FMT_S16P
	case audio.Bits_S32LE, audio.Bits_S24LE:
		return C.AV_SAMPLE_FMT_S32P
	case audio.Bits_32LEF:
		return C.AV_SAMPLE_FMT_FLTP
	case audio.Bits_64LEF:
		return C.AV_SAMPLE_FMT_DBLP
	}
	return C.AV_SAMPLE_FMT_NONE
}

func ChannelsFromLayout(layout uint64) (m audio.Layout, index audio.ChannelIndex) {
	acSlice := []audio.Channel{}

	index = m.ChannelIndex()

	for i := int8(0); i < 64; i++ {
		ch := C.av_channel_layout_extract_channel(C.uint64_t(layout), C.int(i))
		ac := channelFromAV(ch)
		if !ac.IsValid() {
			continue
		}
		index[ac] = i
		acSlice = append(acSlice, ac)
	}

	return audio.NewLayout(acSlice...), index
}

func AVLayoutFromLayout(layout audio.Layout) uint64 {
	av := C.uint64_t(0)
	for _, ch := range layout.ChannelMask.Slice() {
		av |= avFromChannel(ch)
	}
	return uint64(av)
}
