package ffmpeg

/*
#cgo pkg-config: libavformat libavutil libswresample libavcodec
#include "decode.c"
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/ffmpeg/avutil"
)

func New(ofh stream.FileStreamerOpenFileHandler) stream.FileStreamer {

	return &AVFormatContext{
		openedHandler: ofh,
	}
}

type AVFormatContext struct {
	openedHandler stream.FileStreamerOpenFileHandler
	format        *audio.Format
	fileName      string
	pause         bool
	finished      bool

	ctx       *C.GOAVDecoder
	outputFmt *audio.Format

	lastDecodeSize int
	posPerCh       int // 每次读取后，buffer中还剩下的每声道的samples数量
	pos            int // 当前已解码的位置

	lock sync.Mutex
}

func (c *AVFormatContext) OpenFile(fileName string) (err error) {
	c.fileName = fileName
	cFileName := C.CString(fileName)

	defer func() {
		C.free(unsafe.Pointer(cFileName))
		if err != nil {
			c.Close()
		}
	}()

	c.Close()

	rate := C.int(0)
	format := C.enum_AVSampleFormat(C.AV_SAMPLE_FMT_NONE)
	channels := C.int(0)

	ret := C.go_init(&c.ctx, cFileName, &rate, &channels, &format)

	if ret < 0 {
		switch ret {
		case -1:
			return &avutil.AllocError{Size: 0}
		}
		return fmt.Errorf("unknown error")
	}

	bit := avutil.BitsFromAV(format)
	c.format = &audio.Format{
		SampleRate: audio.NewAudioRate(int(rate)),
		Layout:     avutil.ChannelsFromLayout(uint64(C.av_get_default_channel_layout(channels))),
		SampleBits: bit,
	}

	if !c.format.IsValid() {
		return fmt.Errorf("file format invalid: %s", c.format.String())
	}

	if c.outputFmt == nil {
		c.outputFmt = &audio.Format{}
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

func (c *AVFormatContext) Name() string {
	return "ffmpeg"
}

func (c *AVFormatContext) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (c *AVFormatContext) AudioFormat() *audio.Format {
	return c.format
}

func (c *AVFormatContext) SetFormat(format *audio.Format) {
	if !format.SampleRate.IsValid() || !format.SampleBits.IsValid() || format.Layout.Count == 0 {
		return
	}
	*c.outputFmt = *format
	c.initOutputFormat()
}

func (c *AVFormatContext) initOutputFormat() error {
	outputFmt := avutil.FormatFromBits(audio.AudioBits_64LEF)
	rate := C.int(c.outputFmt.SampleRate.ToInt())
	chs := C.int(c.outputFmt.Layout.Count)

	ret := C.go_init_resample(c.ctx, rate, chs, outputFmt)
	if ret < 0 {
		return avutil.NewErrorFromCCode(int(ret))
	}
	return nil
}

func (c *AVFormatContext) Duration() time.Duration {
	if c.ctx == nil {
		return 0
	}
	return time.Duration(c.ctx.duration) * time.Second
}

func (c *AVFormatContext) CurrentFile() string {
	return c.fileName
}

func (c *AVFormatContext) TotalDuration() time.Duration {
	if c.ctx == nil || c.ctx.formatCtx == nil {
		return 0
	}
	return time.Duration(c.ctx.formatCtx.duration) * time.Microsecond
}

func (c *AVFormatContext) Len() int {
	if c.ctx == nil || c.ctx.stream == nil {
		return 0
	}

	return int(c.ctx.stream.nb_frames)
}

func (c *AVFormatContext) Position() int {
	if c.ctx == nil || c.ctx.avFrame == nil {
		return 0
	}

	return int(c.ctx.avFrame.pkt_pos)
}

func (c *AVFormatContext) Pause(p bool) {
	c.pause = p
}

func (c *AVFormatContext) IsPaused() bool {
	return c.pause
}

func (c *AVFormatContext) decode() (n int, err error) {
	ret := C.go_decode(c.ctx)
	if ret < 0 {
		err = avutil.NewErrorFromCCode(int(ret))
	} else {
		n = int(ret)
	}
	return
}

func (c *AVFormatContext) Stream(samples *stream.Samples) {
	if c == nil || c.format == nil || c.ctx == nil {
		return
	}
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

	samples.Format.Mixed(c.format)

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.pause || c.finished {
		nbSamples = samples.BeZero()
		return
	}

	for i = 0; i < samples.Size; i++ {
		if c.posPerCh >= c.lastDecodeSize {
			if c.lastDecodeSize, err = c.decode(); err != nil {
				if c.ctx.finished > C.int(0) {
					c.finished = true
				}
				// 余下数据置零
				samples.BeZeroLeft(i)
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

func (c *AVFormatContext) Seek(p time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	pos := p * time.Duration(c.format.SampleRate) / time.Second
	ret := C.go_seek(c.ctx, C.int(pos))
	if ret < 0 {
		return fmt.Errorf("seek to '%d' failed", pos)
	}
	c.pause = false
	c.finished = false

	return nil
}

func (c *AVFormatContext) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.finished = false
	c.lastDecodeSize = 0
	c.posPerCh = 0
	c.pause = true
	C.go_free(&c.ctx)
	c.ctx = nil
	return nil
}

func (c *AVFormatContext) Debug(d bool) {
	if c.ctx == nil {
		return
	}
	if d {
		c.ctx.debug = C.int(1)
	} else {
		c.ctx.debug = C.int(0)
	}
}
