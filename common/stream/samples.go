package stream

import (
	"fmt"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/utils"
)

// 对应ffmpeg中的planar类型
type Samples struct {
	NbSamples     int // 每声道样本数量，恒等于 Data 的 第二维 数组大小
	Format        audio.Format
	Data          [][]float64 // 第二维数据是指向 Buffer 的 unsafePoint 数组
	Buffer        []byte
	RawData       [][]byte // 第二维数据是指向 Buffer 的 unsafePoint 数组
	LastErr       error    // 最近一次处理的错误码
	LastNbSamples int      // 最近一次处理后剩余的每声道样本数量
	HasData       bool     // 暂停等原因缓存中数据是填充的零

	channelIndex [32]int
}

// Buffer 总大小
func (s *Samples) TotalSize() int {
	// return s.Format.AllSamplesSize(s.NbSamples)
	return len(s.Buffer)
}

// 总样本数量
func (s *Samples) AllSamplesCount() int {
	return s.Format.Layout.Count * s.NbSamples
}

// 每声道的字节数
func (s *Samples) SamplesSize() int {
	return s.Format.SamplesSize(s.NbSamples)
}

// 每声道的字节数
func (s *Samples) LastSamplesSize() int {
	return s.Format.SamplesSize(s.LastNbSamples)
}

func (s *Samples) SetFormat(f audio.Format) {
	var (
		i int
	)
	s.Format = f

	for i = 0; i < int(audio.Channel_MAX); i++ {
		s.channelIndex[i] = -1
	}
	for i, ch := range f.Channels() {
		s.channelIndex[ch] = i
	}
}

func (s *Samples) BeZero() {
	s.HasData = false
	s.BeZeroLeft(0)
}

func (s *Samples) BeZeroLeft(j int) {
	for i := j; i < len(s.Buffer); i++ {
		s.Buffer[i] = 0
	}

	s.LastNbSamples = 0
	s.LastErr = nil
}

func (s *Samples) ChannelBytes(ch int) []byte {
	planar := s.RawData[ch]
	return planar[:s.LastSamplesSize()]
}

func (s *Samples) ChannelsCountBySlice(src []audio.Channel) (c int) {
	for _, ch := range src {
		if s.channelIndex[ch] >= 0 {
			c++
		}
	}
	return
}

func (s *Samples) MixChannel(p []float64, src []audio.Channel) int {
	if s.Format.SampleBits != audio.Bits_DEFAULT {
		return 0
	}
	i := 0
	size := s.LastSamplesSize()

	for _, ch := range src {
		i = s.channelIndex[ch]
		if i < 0 {
			continue
		}
		srcS := s.Data[i][:size]
		for i = 0; i < size; i++ {
			p[i] += srcS[i]
		}
	}

	return size
}

// 从 planar 格式转换至 packed 格式
func (s *Samples) PackedBytes() []byte {
	var (
		p        = make([]byte, s.LastSamplesSize())
		chs      = s.Format.Layout.Count
		bits     = s.Format.SampleBits.Size()
		i, ch, b int
	)
	for i = 0; i < s.LastNbSamples; i++ {
		for ch = 0; ch < chs; ch++ {
			for b = 0; b < bits; b++ {
				p[i*chs+ch*b+b] = s.RawData[ch][i+b]
			}
		}
	}
	return p
}

// 复制GO内存至C内存
//
// dst 是 ffmpeg 中 planar 格式数据，类型 (uint8_t**)，buf[0]指向第一个样本
//
// dstOneSize 声道数量
//
// dstSize dst 总长度（不包含二维数组的头）
//
// @return 每声道已复制的字节数
func (s *Samples) CopyToCBytes(dst unsafe.Pointer, offset int, dstOneSize int, dstSize int) int {
	var (
		dstS       = utils.MakeSlice[unsafe.Pointer](dst, dstOneSize)
		copied     = 0
		dstCh      []byte
		ch         = 0
		dstTwoSize = dstSize
	)

	if dstOneSize > 1 {
		dstTwoSize = int(uintptr(dstS[1]) - uintptr(dstS[0]))
	}

	for ch = 0; ch < dstOneSize && ch < s.Format.Layout.Count; ch++ {
		dstCh = utils.MakeSlice[byte](dstS[ch], dstTwoSize)
		copied = copy(dstCh, s.RawData[ch])
	}

	return copied
}

// 复制C内存至GO内存
//
// src 是 ffmpeg 中 planar 格式数据，类型 (uint8_t**)，buf[0]指向第一个样本
//
// srcOneSize src 的声道数量
//
// srcTwoSize src 的每声道总样本数
//
// dstOffset 每声道复制时 dst 偏移的样本数
//
// srcOffset 每声道复制时 src 偏移的样本数
//
// @return 每声道已复制的字节数
func (s *Samples) CopyFromCBytes(src unsafe.Pointer, srcOneSize, srcTwoSize, dstOffset, srcOffset int) int {
	var (
		srcS   = utils.MakeSlice[unsafe.Pointer](src, srcOneSize)
		srcCh  []byte
		copied = 0
		ch     = 0
	)

	srcOffset = s.Format.SamplesSize(srcOffset)
	dstOffset = s.Format.SamplesSize(dstOffset)
	srcTwoSize = s.Format.SamplesSize(srcTwoSize)

	if srcOffset > srcTwoSize {
		s.LastErr = fmt.Errorf("src offset %d large than buffer size %d", srcOffset, srcTwoSize)
		return 0
	}
	if dstOffset > s.SamplesSize() {
		s.LastErr = fmt.Errorf("dst offset %d large than buffer size %d", dstOffset, s.SamplesSize())
		return 0
	}

	// C样本数量可能小于 Samples 的
	for ch = 0; ch < srcOneSize && ch < s.Format.Layout.Count; ch++ {
		srcCh = utils.MakeSlice[byte](srcS[ch], srcTwoSize)
		copied = copy(s.RawData[ch][dstOffset:], srcCh[srcOffset:])
	}
	if copied != srcTwoSize {
		s.LastErr = fmt.Errorf("buffer size %d per channel not enough, need %d", s.SamplesSize(), srcTwoSize)
	}

	return copied
}

func NewSamples(samples int, format audio.Format) (s *Samples) {
	// 内部处理格式为
	format.SampleBits = audio.Bits_DEFAULT
	format.Layout = audio.ChannalLayoutMAX

	if !format.IsValid() {
		return nil
	}

	p := make([]byte, samples*format.Size())

	chs := format.Layout.Count

	s = &Samples{
		Buffer:    p,
		Data:      make([][]float64, chs),
		RawData:   make([][]byte, chs),
		NbSamples: samples,
	}

	s.SetFormat(format)

	perChSize := s.SamplesSize()

	for ch := 0; ch < chs; ch++ {
		chBuf := unsafe.Pointer(&s.Buffer[ch*samples*format.SampleBits.Size()])

		s.Data[ch] = utils.MakeSlice[float64](chBuf, samples)
		s.RawData[ch] = utils.MakeSlice[byte](chBuf, perChSize)
	}

	return s
}
