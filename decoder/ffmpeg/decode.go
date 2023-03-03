package ffmpeg

/*
#cgo pkg-config: libavformat libavutil libswresample libavcodec
#include "decode.c"
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/decoder"
)

func newErrorFromCCode(code C.int) error {
	size := C.size_t(256)
	buf := (*C.char)(C.av_mallocz(size))
	defer C.av_free(unsafe.Pointer(buf))

	if C.go_averror_is_eof(code) == 1 {
		return &EofError{}
	}

	var err string
	if C.av_strerror(code, buf, size-1) == 0 {
		err = C.GoString(buf)
	} else {
		err = "Unknown error"
	}
	return &Error{
		code: int(code),
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
		return audio.AudioChannel_FRONT_LEFT
	case C.AV_CH_FRONT_RIGHT:
		return audio.AudioChannel_FRONT_RIGHT
	case C.AV_CH_FRONT_CENTER:
		return audio.AudioChannel_FRONT_CENTER
	case C.AV_CH_LOW_FREQUENCY:
		return audio.AudioChannel_LOW_FREQUENCY
	case C.AV_CH_BACK_LEFT:
		return audio.AudioChannel_BACK_LEFT
	case C.AV_CH_BACK_RIGHT:
		return audio.AudioChannel_BACK_RIGHT
	case C.AV_CH_FRONT_LEFT_OF_CENTER:
		return audio.AudioChannel_FRONT_LEFT_OF_CENTER
	case C.AV_CH_FRONT_RIGHT_OF_CENTER:
		return audio.AudioChannel_FRONT_RIGHT_OF_CENTER
	case C.AV_CH_BACK_CENTER:
		return audio.AudioChannel_BACK_CENTER
	case C.AV_CH_SIDE_LEFT:
		return audio.AudioChannel_SIDE_LEFT
	case C.AV_CH_SIDE_RIGHT:
		return audio.AudioChannel_SIDE_RIGHT
	case C.AV_CH_TOP_CENTER:
		return audio.AudioChannel_TOP_CENTER
	case C.AV_CH_TOP_FRONT_LEFT:
		return audio.AudioChannel_TOP_FRONT_LEFT
	case C.AV_CH_TOP_FRONT_CENTER:
		return audio.AudioChannel_TOP_FRONT_CENTER
	case C.AV_CH_TOP_FRONT_RIGHT:
		return audio.AudioChannel_TOP_FRONT_RIGHT
	case C.AV_CH_TOP_BACK_LEFT:
		return audio.AudioChannel_TOP_BACK_LEFT
	case C.AV_CH_TOP_BACK_CENTER:
		return audio.AudioChannel_TOP_BACK_CENTER
	case C.AV_CH_TOP_BACK_RIGHT:
		return audio.AudioChannel_TOP_BACK_RIGHT
	case C.AV_CH_STEREO_LEFT:
	case C.AV_CH_STEREO_RIGHT:
	case C.AV_CH_WIDE_LEFT:
	case C.AV_CH_WIDE_RIGHT:
	}
	return audio.AudioChannel_NONE
}

func bitsFromAV(b C.enum_AVSampleFormat) audio.Bits {
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

func formatFromBits(b audio.Bits) C.enum_AVSampleFormat {
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

func channelsFromLayout(layout C.int64_t) (m audio.ChannelLayout) {
	acSlice := []audio.Channel{}
	for i := 0; i < 64; i++ {
		ch := C.av_channel_layout_extract_channel(C.uint64_t(layout), C.int(i))
		ac := channelFromAV(ch)
		if !ac.IsValid() {
			continue
		}
		acSlice = append(acSlice, ac)
	}
	return audio.NewChannelLayout(acSlice)
}

func New(ofh decoder.FileStreamerOpenFileHandler) decoder.FileStreamer {

	return &avFormatContext{
		openedHandler: ofh,
	}
}

type avFormatContext struct {
	openedHandler decoder.FileStreamerOpenFileHandler
	format        *audio.Format
	fileName      string
	pause         bool

	ctx       *C.GOAVDecoder
	outputFmt *audio.Format

	lastDecodeSize int
	posPerCh       int // 每次读取后，buffer中还剩下的每声道的samples数量
	pos            int // 当前已解码的位置
}

func (c *avFormatContext) OpenFile(fileName string) (err error) {
	c.fileName = fileName
	cFileName := C.CString(fileName)

	defer C.free(unsafe.Pointer(cFileName))

	rate := C.int(0)
	format := C.enum_AVSampleFormat(C.AV_SAMPLE_FMT_NONE)
	channels := C.int(0)

	ret := C.go_init(&c.ctx, cFileName, &rate, &channels, &format)

	if ret < 0 {
		switch ret {
		case -1:
			return &AllocError{0}
		}
	}

	bit := bitsFromAV(format)
	c.format = &audio.Format{
		SampleRate: audio.NewAudioRate(int(rate)),
		Layout:     channelsFromLayout(C.av_get_default_channel_layout(channels)),
		SampleBits: bit,
	}

	if c.outputFmt == nil {
		*c.outputFmt = *c.format
		c.outputFmt.SampleBits = audio.AudioBits_64LEF
	}

	err = c.initOutputFormat()
	if err != nil {
		return err
	}

	c.openedHandler(c, c.format)

	return nil
}

func (c *avFormatContext) Name() string {
	return "ffmpeg"
}

func (c *avFormatContext) Type() decoder.ElementType {
	return decoder.ET_WholeSamples
}

func (c *avFormatContext) AudioFormat() *audio.Format {
	return c.format
}

func (c *avFormatContext) SetFormat(format *audio.Format) {
	if !format.SampleRate.IsValid() || !format.SampleBits.IsValid() || format.Layout.Count == 0 {
		return
	}
	*c.outputFmt = *format
	c.initOutputFormat()
}

func (c *avFormatContext) initOutputFormat() error {
	outputFmt := formatFromBits(audio.AudioBits_64LEF)
	rate := C.int(c.outputFmt.SampleRate.ToInt())
	chs := C.int(c.outputFmt.Layout.Count)

	ret := C.go_init_resample(c.ctx, rate, chs, outputFmt)
	if ret < 0 {
		return newErrorFromCCode(ret)
	}
	return nil
}

func (c *avFormatContext) Duration() time.Duration {
	return 0
}

func (c *avFormatContext) CurrentFile() string {
	return c.fileName
}

func (c *avFormatContext) Len() int {
	return 0
}

func (c *avFormatContext) Position() int {
	return 0
}

func (c *avFormatContext) Pause(p bool) {
	c.pause = p
}

func (c *avFormatContext) IsPaused() bool {
	return c.pause
}

func (c *avFormatContext) decode() (n int, err error) {
	bufSize := C.int(0)
	ret := C.go_decode(c.ctx)
	if ret < 0 {
		err = newErrorFromCCode(ret)
	} else if c.ctx.buffer == nil {
		err = &AllocError{int(bufSize)}
	} else {
		n = int(ret)
	}
	return
}

func (c *avFormatContext) Stream(samples *decoder.Samples) {
	nbSamples := 0
	var (
		chs = c.format.Layout.Count
		pos = uintptr(0)
		i   int
		ch  int
		err error
	)
	defer func() {
		samples.LastSize = nbSamples
		samples.LastErr = err
	}()

	if c.ctx == nil || chs <= 0 {
		return
	}

	if samples.Format.Layout.Count >= chs {
		samples.Format = c.format
	}

	if c.pause {
		for ch = 0; ch < samples.Format.Layout.Count; ch++ {
			for i = 0; i < samples.Size; i++ {
				samples.Buffer[ch][i] = float64(0)
			}
		}
		return
	}

	for i = 0; i < samples.Size; i++ {
		if c.posPerCh >= c.lastDecodeSize {
			if c.lastDecodeSize, err = c.decode(); err != nil {
				return
			}
			c.posPerCh = 0
		}
		pos = uintptr((c.posPerCh * 8 * chs))
		for ch = 0; ch < samples.Format.Layout.Count && ch < chs; ch++ {
			samples.Buffer[ch][i] = *(*float64)(unsafe.Pointer(uintptr(unsafe.Pointer(c.ctx.buffer)) + pos))
			pos += uintptr(8)
		}
		nbSamples++
		c.posPerCh++
	}
}

func (c *avFormatContext) Seek(p time.Duration) error {
	pos := p * time.Duration(c.format.SampleRate) / time.Second

	ret := C.go_seek(c.ctx, C.int(pos))
	if ret < 0 {
		return fmt.Errorf("seek to '%d' failed", pos)
	}
	return nil
}

func (c *avFormatContext) Close() error {
	c.pause = true
	C.go_free(&c.ctx)
	c.ctx = nil
	return nil
}
