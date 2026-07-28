package avutil

/*
#cgo pkg-config: libavutil libavformat

#include "avutil.c"
*/
import "C"
import (
	"errors"
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
	var eofError *EofError
	ok := errors.As(err, &eofError)
	return ok
}

func ChannelFromAV(ch int) audio.Channel {
	switch C.enum_AVChannel(ch) {
	case C.AV_CHAN_FRONT_LEFT:
		return audio.ChannelFrontLeft
	case C.AV_CHAN_FRONT_RIGHT:
		return audio.ChannelFrontRight
	case C.AV_CHAN_FRONT_CENTER:
		return audio.ChannelFrontCenter
	case C.AV_CHAN_LOW_FREQUENCY:
		return audio.ChannelLowFrequency
	case C.AV_CHAN_BACK_LEFT:
		return audio.ChannelBackLeft
	case C.AV_CHAN_BACK_RIGHT:
		return audio.ChannelBackRight
	case C.AV_CHAN_FRONT_LEFT_OF_CENTER:
		return audio.ChannelFrontLeftOfCenter
	case C.AV_CHAN_FRONT_RIGHT_OF_CENTER:
		return audio.ChannelFrontRightOfCenter
	case C.AV_CHAN_BACK_CENTER:
		return audio.ChannelBackCenter
	case C.AV_CHAN_SIDE_LEFT:
		return audio.ChannelSideLeft
	case C.AV_CHAN_SIDE_RIGHT:
		return audio.ChannelSideRight
	// case C.AV_CHAN_TOP_CENTER:
	// 	return audio.Channel_TOP_CENTER
	case C.AV_CHAN_TOP_FRONT_LEFT:
		return audio.ChannelTopFrontLeft
	case C.AV_CHAN_TOP_FRONT_CENTER:
		return audio.ChannelTopFrontCenter
	case C.AV_CHAN_TOP_FRONT_RIGHT:
		return audio.ChannelTopFrontRight
	case C.AV_CHAN_TOP_BACK_LEFT:
		return audio.ChannelTopBackLeft
	case C.AV_CHAN_TOP_BACK_CENTER:
		return audio.ChannelTopBackCenter
	case C.AV_CHAN_TOP_BACK_RIGHT:
		return audio.ChannelTopBackRight
	case C.AV_CHAN_STEREO_LEFT:
	case C.AV_CHAN_STEREO_RIGHT:
	case C.AV_CHAN_WIDE_LEFT:
	case C.AV_CHAN_WIDE_RIGHT:
	}
	return audio.ChannelNone
}

func avFromChannel(ch audio.Channel) uint64 {
	switch ch {
	case audio.ChannelFrontLeft:
		return C.AV_CHAN_FRONT_LEFT
	case audio.ChannelFrontRight:
		return C.AV_CHAN_FRONT_RIGHT
	case audio.ChannelFrontCenter:
		return C.AV_CHAN_FRONT_CENTER
	case audio.ChannelLowFrequency:
		return C.AV_CHAN_LOW_FREQUENCY
	case audio.ChannelBackLeft:
		return C.AV_CHAN_BACK_LEFT
	case audio.ChannelBackRight:
		return C.AV_CHAN_BACK_RIGHT
	case audio.ChannelFrontLeftOfCenter:
		return C.AV_CHAN_FRONT_LEFT_OF_CENTER
	case audio.ChannelFrontRightOfCenter:
		return C.AV_CHAN_FRONT_RIGHT_OF_CENTER
	case audio.ChannelBackCenter:
		return C.AV_CHAN_BACK_CENTER
	case audio.ChannelSideLeft:
		return C.AV_CHAN_SIDE_LEFT
	case audio.ChannelSideRight:
		return C.AV_CHAN_SIDE_RIGHT
	// case audio.Channel_TOP_CENTER:
	// 	return C.AV_CHAN_TOP_CENTER
	case audio.ChannelTopFrontLeft:
		return C.AV_CHAN_TOP_FRONT_LEFT
	case audio.ChannelTopFrontCenter:
		return C.AV_CHAN_TOP_FRONT_CENTER
	case audio.ChannelTopFrontRight:
		return C.AV_CHAN_TOP_FRONT_RIGHT
	case audio.ChannelTopBackLeft:
		return C.AV_CHAN_TOP_BACK_LEFT
	case audio.ChannelTopBackCenter:
		return C.AV_CHAN_TOP_BACK_CENTER
	case audio.ChannelTopBackRight:
		return C.AV_CHAN_TOP_BACK_RIGHT
	default:
		return 0
	}
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
	default:
		return C.AV_SAMPLE_FMT_NONE
	}
}

func AVLayoutFromLayout(layout audio.Layout) uint64 {
	var (
		mask uint64
	)
	for _, ch := range layout.ChannelMask.Slice() {
		mask |= 1 << avFromChannel(ch)
	}
	return mask
}
