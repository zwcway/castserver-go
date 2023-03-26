package stream

import (
	"fmt"
	"unsafe"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/utils"
)

type channelIndexMax [audio.Channel_MAX]int

// 对应ffmpeg中的planar类型
type Samples struct {
	NbSamples     int          // 每声道样本数量，恒等于 Data 的 第二维 数组大小
	Format        audio.Format // 当前样本格式
	Data          [][]float64  // 第二维数据是指向 Buffer 的 unsafePoint 数组
	buffer        []byte
	RawData       [][]byte     // 第二维数据是指向 Buffer 的 unsafePoint 数组
	fmt           audio.Format // buffer 申请的实际格式
	LastErr       error        // 最近一次处理的错误码
	LastNbSamples int          // 最近一次处理后剩余的每声道样本数量

	autoSize     bool
	channelIndex channelIndexMax
}

// Buffer 总大小
func (s *Samples) TotalSize() int {
	// return s.Format.AllSamplesSize(s.NbSamples)
	return len(s.buffer)
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
	if s.Format.Equal(&f) {
		return
	}
	s.setFormat(f, true)
}

func (s *Samples) setFormat(f audio.Format, layout bool) {
	var (
		i int
	)

	for i = 0; i < int(audio.Channel_MAX); i++ {
		s.channelIndex[i] = -1
	}
	if layout {
		s.Format = f
	} else {
		f.Layout = s.Format.Layout
		s.Format = f
	}
	for i, ch := range s.Format.Channels() {
		s.channelIndex[ch] = i
	}
}

func (s *Samples) LessThan(r *Samples) bool {
	if s == nil {
		return false
	}
	if r == nil {
		return false
	}
	if s.Format.Layout.Count < r.Format.Layout.Count {
		return true
	}
	if s.Format.Size() < r.Format.Size() {
		return true
	}
	if s.NbSamples < r.NbSamples {
		return true
	}

	return false
}

func (s *Samples) ResetData() {
	s.BeZeroLeft(0)

	s.LastNbSamples = 0
	s.LastErr = nil
}

func (s *Samples) ResetFormat() {
	s.Format = s.fmt
}

func (s *Samples) ResetAll() {
	s.ResetData()
	s.ResetFormat()
}

func (s *Samples) BeZeroLeft(j int) {
	l := len(s.buffer)
	for ; j < l; j++ {
		s.buffer[j] = 0
	}
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

func (src *Samples) MixChannelMap(dst *Samples, dstOffset, srcOffset int) (mixed int) {
	if src.Format.Layout.Count == 1 && dst.Format.Layout.Count > 1 {
		left := dst.channelIndex[audio.Channel_FRONT_LEFT]
		if left >= 0 {
			mixed = src.mixChannel(dst, left, 0, dstOffset, srcOffset)
		}
		right := dst.channelIndex[audio.Channel_FRONT_RIGHT]
		if right >= 0 {
			mixed = src.mixChannel(dst, right, 0, dstOffset, srcOffset)
		}
		return
	}
	if src.Format.Layout.Count > 1 && dst.Format.Layout.Count == 1 {
		left := src.channelIndex[audio.Channel_FRONT_LEFT]
		if left >= 0 {
			mixed = src.mixChannel(dst, 0, left, dstOffset, srcOffset)
		}
		right := src.channelIndex[audio.Channel_FRONT_RIGHT]
		if right >= 0 {
			mixed = src.mixChannel(dst, 0, right, dstOffset, srcOffset)
		}
		return
	}

	return src.MixChannels(dst, src.Format.Channels(), dstOffset, srcOffset)
}

func (src *Samples) mixChannel(dst *Samples, dstCh, srcCh int, dstOffset, srcOffset int) int {
	var (
		i = dstOffset
		j = srcOffset
	)

	for i < dst.NbSamples && j < src.LastNbSamples {
		dst.Data[dstCh][i] += src.Data[srcCh][j]
		i++
		j++
	}

	return j - srcOffset
}

func (src *Samples) Mix(dst *Samples, dstOffset int, srcOffset int) int {
	return src.MixChannels(dst, src.Format.Channels(), dstOffset, srcOffset)
}

func (src *Samples) MixChannels(dst *Samples, srcChs []audio.Channel, dstOffset int, srcOffset int) int {
	if src.Format.SampleBits != audio.Bits_DEFAULT || dst == nil || dst.Format.SampleBits != audio.Bits_DEFAULT {
		return 0
	}
	var (
		i     = 0
		j     = 0
		mixed = 0
	)

	for _, ch := range srcChs {
		if !ch.IsValid() {
			continue
		}

		i = src.channelIndex[ch]
		if i < 0 {
			continue
		}
		j = dst.channelIndex[ch]
		if j < 0 {
			continue
		}

		i = src.mixChannel(dst, j, i, dstOffset, srcOffset)
		if mixed < i {
			mixed = i
		}
	}

	return mixed
}

func (s *Samples) ChannelIndex(ch audio.Channel) int {
	if !ch.IsValid() {
		return -1
	}
	return s.channelIndex[ch]
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

func sampleSize(samples *int, format *audio.Format) bool {
	autoSize := *samples == 0
	if autoSize {
		// 取 10ms
		*samples = format.SampleRate.ToInt() * 10 / 1000
	}
	return autoSize
}

func NewSamples(samples int, format audio.Format) (s *Samples) {
	if !format.IsValid() {
		return nil
	}
	autoSize := sampleSize(&samples, &format)

	p := make([]byte, samples*format.Size())

	s = &Samples{
		Format: format,
	}
	reuseSamples(s, p, format)
	s.setFormat(format, true)
	s.autoSize = autoSize

	return
}

func NewSamplesCopy(p []byte, format audio.Format) (s *Samples) {
	// 向下取整
	samples := len(p) / format.Size()
	s = NewSamples(samples, format)
	copy(s.buffer, p)
	s.LastNbSamples = samples
	return
}

func ReuseSamples(p []byte, format audio.Format) (s *Samples) {
	s = &Samples{
		Format: format,
	}
	reuseSamples(s, p, format)
	s.setFormat(format, true)
	s.LastNbSamples = s.NbSamples

	return s
}

func reuseSamples(s *Samples, p []byte, format audio.Format) {
	chs := format.Layout.Count
	samples := len(p) / format.Size()

	s.buffer = p
	s.Data = make([][]float64, chs)
	s.RawData = make([][]byte, chs)
	s.NbSamples = samples
	s.Format = format
	s.fmt = format

	perChSize := s.SamplesSize()

	for ch := 0; ch < chs; ch++ {
		chBuf := unsafe.Pointer(&s.buffer[ch*samples*format.SampleBits.Size()])

		s.Data[ch] = utils.MakeSlice[float64](chBuf, samples)
		s.RawData[ch] = utils.MakeSlice[byte](chBuf, perChSize)
	}
}

func (s *Samples) resizeSample(samples int, format audio.Format, resizeLayout bool) {
	autoSize := sampleSize(&samples, &format)

	if resizeLayout {
		if !s.fmt.LessThan(&format) && s.NbSamples >= samples {
			s.setFormat(format, resizeLayout)
			return
		}
	} else {
		if !s.fmt.SampleLessThan(&format) && s.NbSamples >= samples {
			s.setFormat(format, resizeLayout)
			return
		}
	}

	if !autoSize {
		autoSize = s.autoSize
	}
	p := make([]byte, samples*format.Size())
	copy(p, s.buffer)
	reuseSamples(s, p, format)
	s.setFormat(format, true)
	s.autoSize = autoSize
}

func (s *Samples) ResizeSamplesExceptLayout(samples int, format audio.Format) {
	s.resizeSample(samples, format, false)
}

func (s *Samples) ResizeSamplesOrNot(samples int, format audio.Format) {
	s.resizeSample(samples, format, true)
}

func (s *Samples) ChannelSamples(ch audio.Channel) *Samples {
	si := s.channelIndex[ch]
	if si < 0 {
		return nil
	}

	format := s.Format
	format.Layout = audio.NewChannelLayout(ch)

	ns := &Samples{
		Format:    format,
		buffer:    s.buffer,
		Data:      make([][]float64, 1),
		RawData:   make([][]byte, 1),
		NbSamples: s.NbSamples,
		fmt:       format,
	}
	ns.setFormat(format, true)

	// 将任意的声道都映射至第一个
	for i := 0; i < int(audio.Channel_MAX); i++ {
		s.channelIndex[i] = 0
	}

	ns.Data[0] = s.Data[si]
	ns.RawData[0] = s.RawData[si]

	return ns
}
