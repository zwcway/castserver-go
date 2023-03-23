package resample

/*
#cgo pkg-config: libavutil libswresample
#include "resample.c"

*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/ffmpeg/avutil"
)

type Resample struct {
	swrCtx *C.GOResample

	inFormat  audio.Format
	inBufSize int

	outFormat  audio.Format
	outBufSize int
}

func (r *Resample) SetIn(ifmt audio.Format) error {
	if ifmt.Equal(&r.inFormat) {
		return nil
	}
	r.inFormat = ifmt

	return r.init()
}

func (r *Resample) SetOut(ofmt audio.Format) error {
	if ofmt.Equal(&r.outFormat) {
		return nil
	}
	r.outFormat = ofmt

	return r.init()
}

func (r *Resample) Inited() bool {
	return r.swrCtx != nil
}

func (r *Resample) init() error {
	r.outFormat.InitFrom(r.inFormat)

	if !r.inFormat.IsValid() || !r.outFormat.IsValid() {
		return errors.New("set format first")
	}
	avInLayout := avutil.AVLayoutFromChannelLayout(r.inFormat.Layout)
	avOutLayout := avutil.AVLayoutFromChannelLayout(r.outFormat.Layout)

	ret := C.go_swr_init(&r.swrCtx,
		C.int(r.inFormat.SampleRate.ToInt()), C.int64_t(avInLayout), avutil.AVFormatFromBits(r.inFormat.SampleBits),
		C.int(r.outFormat.SampleRate.ToInt()), C.int64_t(avOutLayout), avutil.AVFormatFromBits(r.outFormat.SampleBits))

	if ret < 0 {
		r.Close()
		return avutil.NewErrorFromCCode(int(ret))
	}

	return nil
}

func (r *Resample) Close() {
	if r.swrCtx != nil {
		C.go_swr_free(&r.swrCtx)
	}
}

func bufferSize(samplesPerChannel int, format *audio.Format) int {
	return format.Layout.Count * samplesPerChannel * format.SampleBits.Size()
}

func (r *Resample) Stream(samples *stream.Samples) error {
	if r.swrCtx == nil {
		return errors.New("need init")
	}
	if samples.LastNbSamples == 0 {
		return nil
	}
	if err := r.SetIn(samples.Format); err != nil {
		samples.LastErr = &avutil.AllocError{Size: r.inBufSize}
		return err
	}

	ret := C.malloc_in_buffer(r.swrCtx, C.int(samples.LastNbSamples))
	if ret < 0 {
		samples.LastErr = avutil.NewErrorFromCCode(int(ret))
		r.inBufSize = 0
		return samples.LastErr
	}

	// 复制Go内存至C内存
	samples.CopyToCBytes(unsafe.Pointer(r.swrCtx.in_buffer), 0,
		r.inFormat.Layout.Count, samples.LastSamplesSize())

	// inBuffer 与 outBuffer 都是二维数组，uint8_t**
	ret = C.go_convert(r.swrCtx, C.int(samples.LastNbSamples))

	if ret < 0 {
		samples.LastErr = avutil.NewErrorFromCCode(int(ret))
		return samples.LastErr
	}

	samples.ResizeSamples(int(ret), r.outFormat)

	// 复制C内存至Go内存
	samples.CopyFromCBytes(unsafe.Pointer(r.swrCtx.out_buffer), r.outFormat.Layout.Count, int(ret), 0, 0)

	return nil
}
