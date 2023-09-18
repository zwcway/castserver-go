package ffmpeg

/*
#cgo pkg-config: libavformat libavutil libswresample libavcodec
#include "decode.c"
*/
import "C"
import (
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/playlist"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/ffmpeg/avutil"
)

func New(outFormat audio.Format) stream.FileStreamer {
	return &AVFormatContext{
		outputFmt: outFormat,
	}
}

func AudioInfo(f string, ai *playlist.AudioInfo) (err error) {
	var (
		cFileName = C.CString(f)
		ctx       *C.GOAVDecoder
		tag       *C.AVDictionaryEntry
		cEmptyStr = C.CString("")
	)
	defer func() {
		C.free(unsafe.Pointer(cFileName))
		C.free(unsafe.Pointer(cEmptyStr))
		if err != nil {
			C.go_free(&ctx)
		}
	}()

	rate := C.int(0)
	format := C.enum_AVSampleFormat(C.AV_SAMPLE_FMT_NONE)
	channels := C.int(0)

	ret := C.go_init(&ctx, cFileName, &rate, &channels, &format)

	if ret < 0 {
		return avutil.NewErrorFromCCode(int(ret))
	}

	bit := avutil.BitsFromAVFormat(format)
	ai.Format = audio.Format{
		Sample: audio.Sample{
			Rate: audio.NewAudioRate(int(rate)),
			Bits: bit,
		},
	}
	ai.Format.Layout, _ = avutil.ChannelsFromLayout(uint64(ctx.codecCtx.channel_layout))

	ai.Url = f
	ai.Duration = time.Duration(ctx.formatCtx.duration) * time.Microsecond
	ai.Position = time.Duration(ctx.duration) * time.Second

	for {
		tag = C.av_dict_get(ctx.formatCtx.metadata, cEmptyStr, tag, C.AV_DICT_IGNORE_SUFFIX)
		if tag == nil {
			break
		}
		key := strings.ToLower(C.GoString(tag.key))
		val := C.GoString(tag.value)
		switch key {
		case "title":
			ai.Title = val
		case "artist":
			ai.Artist = val
		}
	}
	return nil
}

type AVFormatContext struct {
	format   audio.Format
	fileName string
	pause    bool
	finished bool

	ctx                *C.GOAVDecoder
	channelIndex       audio.ChannelIndex
	outputFmt          audio.Format
	outBufferSize      int //
	outBufferNbSamples int

	lastDecodeNbSamples int
	posDecodeNbSamples  int // 每次读取后，buffer中还剩下的每声道的samples数量
	pos                 int // 当前已解码的位置

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

	bit := avutil.BitsFromAVFormat(format)
	c.format = audio.Format{
		Sample: audio.Sample{
			Rate: audio.NewAudioRate(int(rate)),
			Bits: bit,
		},
	}
	c.format.Layout, c.channelIndex = avutil.ChannelsFromLayout(uint64(c.ctx.codecCtx.channel_layout))

	if !c.format.IsValid() {
		return fmt.Errorf("file format invalid: %s", c.format.String())
	}

	// 如果未初始化 outputFmt，使用 当前音频 格式初始化
	c.outputFmt.InitFrom(c.format)
	// c.outputFmt.SampleBits = audio.Bits_DEFAULT

	err = c.SetOutFormat(c.outputFmt)
	if err != nil {
		return err
	}

	stream.BusSourceFormatChanged.Dispatch(c, &c.format, c.channelIndex)

	return nil
}

func (c *AVFormatContext) Name() string {
	return "ffmpeg"
}

func (c *AVFormatContext) Type() stream.ElementType {
	return stream.ET_WholeSamples
}

func (c *AVFormatContext) AudioFormat() audio.Format {
	return c.format
}

func (c *AVFormatContext) ChannelIndex() audio.ChannelIndex {
	return c.channelIndex
}

func (c *AVFormatContext) OutFormat() audio.Format {
	return c.outputFmt
}

func (c *AVFormatContext) SetOutFormat(f audio.Format) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	f.InitFrom(c.format)

	// 不改变声道数量
	// f.Layout = c.format.Layout

	c.outputFmt = f

	err := c.initOutputFormat()
	if err != nil {
		return err
	}

	return nil
}

func (c *AVFormatContext) initOutputFormat() error {
	outputFmt := avutil.AVFormatFromBits(c.outputFmt.Bits)
	rate := C.int(c.outputFmt.Rate.ToInt())
	chs_layout := avutil.AVLayoutFromLayout(c.outputFmt.Layout)

	ret := C.go_init_resample(c.ctx, rate, C.int64_t(chs_layout), outputFmt)
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

func (c *AVFormatContext) IsPlaying() bool {
	return !c.pause && !c.finished
}

func (c *AVFormatContext) CanRemove() bool {
	return false
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

func (c *AVFormatContext) SetPause(p bool) {
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
	c.outBufferSize = int(c.ctx.swrCtx.out_buf_size)
	c.outBufferNbSamples = int(c.ctx.nb_samples)
	return
}

func (c *AVFormatContext) Stream(samples *stream.Samples) {
	if c == nil || !c.format.IsValid() || c.ctx == nil {
		return
	}
	nbSamples := 0
	var (
		chs = int(c.format.Count)
		err error
	)

	defer func() {
		samples.LastNbSamples = nbSamples
		samples.LastErr = err
	}()

	if c.ctx == nil || chs <= 0 {
		return
	}

	samples.SetFormatAndIndex(c.outputFmt, c.channelIndex)

	if c.pause || c.finished {
		// samples.BeZero()
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	for nbSamples < samples.RequestNbSamples {
		if c.posDecodeNbSamples >= c.lastDecodeNbSamples {
			if c.lastDecodeNbSamples, err = c.decode(); err != nil {
				if c.ctx.finished > C.int(0) {
					c.finished = true
					c.pause = true
				}
				// 余下数据置零
				samples.BeZeroLeft(samples.RequestNbSamples - nbSamples)
				return
			}
			c.posDecodeNbSamples = 0
		}

		copied := samples.CopyFromCBytes(unsafe.Pointer(c.ctx.buffer), chs, c.outBufferNbSamples,
			nbSamples, c.posDecodeNbSamples)

		copied = samples.Format.SamplesCount(copied)

		nbSamples += copied
		c.posDecodeNbSamples += copied
	}
}

func (c *AVFormatContext) Seek(p time.Duration) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	pos := p * time.Duration(c.format.Rate) / time.Second
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
	c.lastDecodeNbSamples = 0
	c.posDecodeNbSamples = 0
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
