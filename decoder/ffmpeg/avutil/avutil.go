package avutil

/*
#cgo pkg-config: libavutil libavformat

#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
#include <libavutil/samplefmt.h>

static int go_averror_is_eof(int code)
{
    return code == AVERROR_EOF;
}
static void *go_malloc(int size)
{
    return av_malloc(size);
}
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

func ChannelFromAV(ch C.uint64_t) audio.Channel {
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
	case C.AV_CH_TOP_CENTER:
		return audio.Channel_TOP_CENTER
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

func BitsFromAV(b C.enum_AVSampleFormat) audio.Bits {
	switch b {
	case C.AV_SAMPLE_FMT_U8, C.AV_SAMPLE_FMT_U8P:
		return audio.AudioBits_U8
	case C.AV_SAMPLE_FMT_S16, C.AV_SAMPLE_FMT_S16P:
		return audio.AudioBits_S16LE
	case C.AV_SAMPLE_FMT_S32, C.AV_SAMPLE_FMT_S32P:
		return audio.AudioBits_S32LE
	case C.AV_SAMPLE_FMT_FLT, C.AV_SAMPLE_FMT_FLTP:
		return audio.AudioBits_32LEF
	case C.AV_SAMPLE_FMT_DBL:
	case C.AV_SAMPLE_FMT_DBLP:
	case C.AV_SAMPLE_FMT_S64:
	case C.AV_SAMPLE_FMT_S64P:
	}
	return audio.AudioBits_NONE
}

func FormatFromBits(b audio.Bits) C.enum_AVSampleFormat {
	switch b {
	case audio.AudioBits_U8:
		return C.AV_SAMPLE_FMT_U8
	case audio.AudioBits_S16LE:
		return C.AV_SAMPLE_FMT_S16
	case audio.AudioBits_S32LE, audio.AudioBits_S24LE:
		return C.AV_SAMPLE_FMT_S32
	case audio.AudioBits_32LEF:
		return C.AV_SAMPLE_FMT_FLT
	case audio.AudioBits_64LEF:
		return C.AV_SAMPLE_FMT_DBL
	}
	return C.AV_SAMPLE_FMT_NONE
}

func ChannelsFromLayout(layout uint64) (m audio.ChannelLayout) {
	acSlice := []audio.Channel{}
	for i := 0; i < 64; i++ {
		ch := C.av_channel_layout_extract_channel(C.uint64_t(layout), C.int(i))
		ac := ChannelFromAV(ch)
		if !ac.IsValid() {
			continue
		}
		acSlice = append(acSlice, ac)
	}
	return audio.NewChannelLayout(acSlice...)
}

// 构建一个新的 C 二维数组 uint8_t*[]
func NewPointerArray(one int, two int) unsafe.Pointer {
	ptrSize := unsafe.Sizeof(uintptr(0))
	cahead := one * int(ptrSize)
	c := (**C.uint8_t)(C.go_malloc(C.int(one*two + cahead)))
	if c == nil {
		return nil
	}
	for ch := 0; ch < one; ch++ {
		ap := (*C.uint8_t)(unsafe.Pointer(uintptr(unsafe.Pointer(c)) + uintptr(ch*two+cahead)))
		ai := (**C.uint8_t)(unsafe.Pointer(uintptr(unsafe.Pointer(c)) + uintptr(ch)*ptrSize))
		*ai = ap
	}
	return unsafe.Pointer(c)
}
