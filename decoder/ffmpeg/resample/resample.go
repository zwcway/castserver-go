package resample

/*
#cgo pkg-config: libavutil libswresample
#include <libswresample/swresample.h>
*/
import "C"
import (
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/castserver-go/decoder/ffmpeg/avutil"
)

type Resample struct {
	swrCtx *C.SwrContext

	inFormat  audio.Format
	inBufSize int
	inBuffer  unsafe.Pointer

	outFormat  audio.Format
	outBufSize int
	outBuffer  unsafe.Pointer
}

func (r *Resample) SetIn(ifmt *audio.Format) {
	if ifmt.Equal(&r.inFormat) {
		return
	}
	r.inFormat = *ifmt

	r.init()
}

func (r *Resample) SetOut(ofmt *audio.Format) {
	if ofmt.Equal(&r.outFormat) {
		return
	}
	r.outFormat = *ofmt

	r.init()
}

func (r *Resample) init() error {

	inChannelLayout := C.av_get_default_channel_layout(C.int(r.inFormat.Layout.Count))
	outChannelLayout := C.av_get_default_channel_layout(C.int(r.outFormat.Layout.Count))

	if r.swrCtx == nil {
		// 初始化转码器
		r.swrCtx = C.swr_alloc()
		if r.swrCtx == nil {
			return &avutil.AllocError{}
		}
	}
	if r.outFormat.SampleBits == audio.AudioBits_NONE {
		r.outFormat.SampleBits = r.inFormat.SampleBits
	}
	if r.outFormat.Layout.Count == 0 {
		r.outFormat.Layout = r.inFormat.Layout
	}
	if r.outFormat.SampleRate == audio.AudioRate_NONE {
		r.outFormat.SampleRate = r.inFormat.SampleRate
	}

	r.swrCtx = C.swr_alloc_set_opts(r.swrCtx,
		outChannelLayout, avutil.FormatFromBits(r.outFormat.SampleBits), C.int(r.outFormat.SampleRate.ToInt()),
		inChannelLayout, avutil.FormatFromBits(r.inFormat.SampleBits), C.int(r.inFormat.SampleRate.ToInt()),
		0, nil)

	ret := C.swr_init(r.swrCtx)
	if ret < 0 {
		r.Close()
		return avutil.NewErrorFromCCode(int(ret))
	}
	return nil
}

func (r *Resample) Close() {
	if r.swrCtx != nil {
		C.swr_free((**C.SwrContext)(unsafe.Pointer(&r.swrCtx)))
	}
	if r.inBuffer != nil {
		C.av_freep(unsafe.Pointer(&r.inBuffer))
	}
	if r.outBuffer != nil {
		C.av_freep(unsafe.Pointer(&r.outBuffer))
	}
}

func inOutSize(samplesPerChannel int, format *audio.Format) int {
	return format.Layout.Count * samplesPerChannel * format.SampleBits.Size()
}

func initBuffer(buf *unsafe.Pointer, bufSize *int, samples *decoder.Samples, format *audio.Format) bool {
	bs := inOutSize(samples.Size, format)

	if *bufSize < bs {
		if *buf != nil {
			C.av_freep(unsafe.Pointer(buf))
		}
		*buf = C.av_malloc(_Ctype_ulonglong(bs))
		if *buf == nil {
			*bufSize = 0
			samples.LastErr = &avutil.AllocError{Size: bs}
			return false
		}
		*bufSize = bs
	}
	return true
}

func (r *Resample) Stream(samples *decoder.Samples) {
	r.SetIn(samples.Format)

	if !initBuffer(&r.inBuffer, &r.inBufSize, samples, &r.inFormat) {
		return
	}
	if !initBuffer(&r.outBuffer, &r.outBufSize, samples, &r.outFormat) {
		return
	}
	// 复制Go内存缓存至C内存缓存
	for ch := 0; ch < r.inFormat.Layout.Count; ch++ {
		for i := 0; i < samples.Size; i++ {
			*(*C.uint8_t)(unsafe.Pointer(uintptr(r.inBufSize) + uintptr(ch*samples.Size+i))) = C.uint8_t(samples.Buffer[ch][i])
		}
	}

	// inBuffer 与 outBuffer 都是二维数组，uint8_t*[]
	ret := C.swr_convert(r.swrCtx,
		(**C.uint8_t)(r.outBuffer), C.int(samples.Size),
		(**C.uint8_t)(r.inBuffer), C.int(samples.Size))

	if ret < 0 {
		samples.LastErr = avutil.NewErrorFromCCode(int(ret))
		return
	}

	samples.Size = int(ret)
	samples.Format = &r.outFormat
}
