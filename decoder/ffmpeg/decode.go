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
	"github.com/zwcway/castserver-go/decoder/ffmpeg/avutil"
)

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
			return &avutil.AllocError{0}
		}
	}

	bit := avutil.BitsFromAV(format)
	c.format = &audio.Format{
		SampleRate: audio.NewAudioRate(int(rate)),
		Layout:     avutil.ChannelsFromLayout(uint64(C.av_get_default_channel_layout(channels))),
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
	outputFmt := avutil.FormatFromBits(audio.AudioBits_64LEF)
	rate := C.int(c.outputFmt.SampleRate.ToInt())
	chs := C.int(c.outputFmt.Layout.Count)

	ret := C.go_init_resample(c.ctx, rate, chs, outputFmt)
	if ret < 0 {
		return avutil.NewErrorFromCCode(int(ret))
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
		err = avutil.NewErrorFromCCode(int(ret))
	} else if c.ctx.buffer == nil {
		err = &avutil.AllocError{Size: int(bufSize)}
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
