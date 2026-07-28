package resample

/*
#cgo pkg-config: libavutil libswresample libavcodec
#include "resample.c"

*/
import "C"
import (
	"errors"
	"sync"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/modules/decoder/ffmpeg/avutil"
)

type Resample struct {
	// mu 保护 swrCtx 与 inFormat/outFormat 的并发访问。
	// Stream 与 SetIn/SetOut 可能在不同 goroutine 中触发（pipeline 调度）。
	mu sync.Mutex

	swrCtx *C.GOResample

	inFormat  audio.Format
	outFormat audio.Format
}

func (r *Resample) SetIn(ifmt audio.Format) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ifmt.Equal(r.inFormat) && r.swrCtx != nil {
		return nil
	}
	r.inFormat = ifmt

	return r.initLocked()
}

func (r *Resample) SetOut(ofmt audio.Format) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ofmt.Equal(r.outFormat) && r.swrCtx != nil {
		return nil
	}
	r.outFormat = ofmt

	return r.initLocked()
}

func (r *Resample) Inited() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.swrCtx != nil
}

func (r *Resample) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closeLocked()
}

// closeLocked 必须在持有 mu 时调用
func (r *Resample) closeLocked() {
	if r.swrCtx != nil {
		C.go_swr_free(&r.swrCtx)
	}
}

func (r *Resample) Stream(samples *stream.Samples) error {
	if samples == nil {
		return nil
	}
	if samples.LastNbSamples == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 分配 C 端输入缓冲区（planar：每个声道独立 buffer）
	ret := C.malloc_in_buffer(r.swrCtx, C.int(samples.LastNbSamples))
	if ret < 0 {
		samples.LastErr = avutil.NewErrorFromCCode(int(ret))
		return samples.LastErr
	}

	// 复制Go内存至C内存
	avutil.CopyToCBytes(unsafe.Pointer(r.swrCtx.in_buffer), samples, int(r.inFormat.Count), samples.LastSamplesSize())

	// inBuffer 与 outBuffer 都是 planar：uint8_t*[chs] 指针数组
	ret = C.go_convert(r.swrCtx, C.int(samples.LastNbSamples))
	if ret < 0 {
		samples.LastErr = avutil.NewErrorFromCCode(int(ret))
		return samples.LastErr
	}

	// 确保有足够空间存储
	samples.Resize(int(ret), r.outFormat)
	samples.LastNbSamples = int(ret)

	// 复制C内存至Go内存
	avutil.CopyFromCBytes(samples, unsafe.Pointer(r.swrCtx.out_buffer), int(r.outFormat.Count), int(ret), 0, 0)

	// 成功路径：清空历史错误
	samples.LastErr = nil
	return nil
}

// initLocked 等价于 init()，但调用方已持有 mu，不会再次加锁。
func (r *Resample) initLocked() error {
	if !r.outFormat.IsValid() {
		r.outFormat.InitFrom(r.inFormat)
	}

	if !r.inFormat.IsValid() || !r.outFormat.IsValid() {
		return errors.New("set format first")
	}
	avInLayout := C.AVChannelLayout{}
	avOutLayout := C.AVChannelLayout{}
	mask := avutil.AVLayoutFromLayout(r.inFormat.Layout)
	C.av_channel_layout_from_mask(&avInLayout, C.uint64_t(mask))
	mask = avutil.AVLayoutFromLayout(r.outFormat.Layout)
	C.av_channel_layout_from_mask(&avOutLayout, C.uint64_t(mask))

	inFmt := avutil.AVFormatFromBits(r.inFormat.Bits)
	outFmt := avutil.AVFormatFromBits(r.outFormat.Bits)

	ret := C.go_swr_init(&r.swrCtx,
		C.int(r.inFormat.Rate.ToInt()), &avInLayout, inFmt,
		C.int(r.outFormat.Rate.ToInt()), &avOutLayout, outFmt)

	if ret < 0 {
		r.closeLocked()
		return avutil.NewErrorFromCCode(int(ret))
	}

	return nil
}
