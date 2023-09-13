package stream

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/utils"
)

// 对应ffmpeg中的planar类型
// TODO 增加所有声道的 Samples，自动同步 Format
type Samples struct {
	RequestNbSamples int          // 请求的每声道样本数量
	Format           audio.Format // 当前样本格式
	Data             [][]float64  // 第二维数据是指向 Buffer 的 unsafePoint 数组
	RawData          [][]byte     // 第二维数据是指向 Buffer 的 unsafePoint 数组
	LastErr          error        // 最近一次处理的错误码
	LastNbSamples    int          // 最近一次处理后剩余的每声道样本数量

	buffer         []byte
	fmt            audio.Format // buffer 申请的实际格式
	autoSize       bool
	channelSamples []*Samples         // 每个声道的样本对象
	ChannelIndex   audio.ChannelIndex // channel和索引的映射表
}

// Buffer 总大小，包含所有声道
func (s *Samples) BufferSize() int {
	// return s.Format.AllSamplesSize(s.nbSamples)
	return len(s.buffer)
}

// 当前总样本数量
func (s *Samples) AllSamplesCount() int {
	return int(s.Format.Layout.Count) * s.LastNbSamples
}

// 每声道中实际可存放的最大字节数
func (s *Samples) SamplesSize() int {
	return len(s.buffer) / int(s.fmt.Count)
}

// 每声道中实际可存放的最大样本数
func (s *Samples) SamplesCount() int {
	return s.SamplesSize() / s.fmt.Bits.Size()
}

// 每声道中请求存放的最大字节数
func (s *Samples) RequestSamplesSize() int {
	return s.Format.SamplesSize(s.RequestNbSamples)
}

// 当前每声道已存入的字节数
func (s *Samples) LastSamplesSize() int {
	return s.Format.SamplesSize(s.LastNbSamples)
}

func (s *Samples) SetRate(f audio.Rate) {
	s.Format.Sample.Rate = f
}

// 在 SourceStreamer 中调用
func (s *Samples) SetFormatAndIndex(f audio.Format, ci *audio.ChannelIndex) {
	if s.Format == f {
		return
	}
	s.Format = f
	s.ChannelIndex = *ci
}

// 更改声道布局
func (s *Samples) SetLayout(f audio.Layout) {
	if s.Format.Count >= f.Count {
		return
	}
	if len(s.Data) < int(f.Count) {
		s.setLayout(f)
	} else {
		s.Format.Layout = f
	}
}

func (s *Samples) setLayout(f audio.Layout) {
	var (
		ch        int
		chs       = int(f.Count)
		samples   = s.RequestNbSamples
		chBuf     unsafe.Pointer
		perChSize int = s.SamplesSize()
		bits      int = s.Format.Sample.Bits.Size()
	)

	s.Format.Layout = f

	s.Data = make([][]float64, chs)
	s.RawData = make([][]byte, chs)

	for ch = 0; ch < chs; ch++ {
		chBuf = unsafe.Pointer(&s.buffer[ch*samples*bits])

		s.Data[ch] = utils.MakeSlice[float64](chBuf, samples)
		s.RawData[ch] = utils.MakeSlice[byte](chBuf, perChSize)
	}

	// s.setChannelIndex()
}

func (s *Samples) SetChannelIndex(ci *audio.ChannelIndex) {
	if ci == nil {
		s.ChannelIndex = *s.Format.ChannelIndex()
		return
	}

	s.ChannelIndex = *ci
}

func (s *Samples) LessThan(r *Samples) bool {
	if s == nil {
		return true
	}
	if r == nil {
		return false
	}

	if len(s.buffer) < len(r.buffer) {
		return true
	}

	return false
}

func (s *Samples) ResetData() {
	s.buffer[0] = byte(0)
	s.buffer[1] = byte(0)
	s.buffer[2] = byte(0)
	s.buffer[4] = byte(0)
	s.buffer[5] = byte(0)
	s.buffer[6] = byte(0)
	s.buffer[7] = byte(0)
	for bp := 8; bp < len(s.buffer); bp <<= 1 {
		copy(s.buffer[bp:], s.buffer[:bp])
	}
	s.RequestNbSamples = len(s.buffer) / int(s.fmt.Count) / s.fmt.Bits.Size()
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
	for _, ch := range s.Data {
		for ; j < len(ch); j++ {
			ch[j] = 0
		}
	}
}

func (s *Samples) ChannelBytes(ch audio.Channel) []byte {
	c := s.ChannelIndex[ch]
	if c < 0 {
		return nil
	}
	planar := s.RawData[c]
	return planar[:s.LastSamplesSize()]
}

// 混合声道，处理声道数量不一致的情况
func (src *Samples) MixChannelMap(dst *Samples, dstOffset, srcOffset int) (mixed int) {
	if sc := src.ChannelIndex[audio.Channel_FRONT_CENTER]; sc >= 0 && src.Format.Layout.Count == 1 && dst.Format.Layout.Count > 1 {
		dc := dst.ChannelIndex[audio.Channel_FRONT_CENTER]
		if dc >= 0 {
			mixed = src.mixChannel(dst, dc, sc, dstOffset, srcOffset)
			return
		}
		// 混合至前左和前右
		dc = dst.ChannelIndex[audio.Channel_FRONT_LEFT]
		if dc >= 0 {
			mixed = src.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		dc = dst.ChannelIndex[audio.Channel_FRONT_RIGHT]
		if dc >= 0 {
			mixed = src.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		return
	}

	if dc := dst.ChannelIndex[audio.Channel_FRONT_CENTER]; dc >= 0 && src.Format.Layout.Count > 1 && dst.Format.Layout.Count == 1 {
		sc := src.ChannelIndex[audio.Channel_FRONT_CENTER]

		if sc >= 0 {
			mixed = src.mixChannel(dst, dc, sc, dstOffset, srcOffset)
			return
		}
		// 从前左和前右混合
		sc = src.ChannelIndex[audio.Channel_FRONT_LEFT]
		if sc >= 0 {
			mixed = src.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		sc = src.ChannelIndex[audio.Channel_FRONT_RIGHT]
		if sc >= 0 {
			mixed = src.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		return
	}

	return src.MixChannels(dst, src.Format.Channels(), dstOffset, srcOffset)
}

func (src *Samples) mixChannel(dst *Samples, dstCh, srcCh int8, dstOffset, srcOffset int) int {
	var (
		i = dstOffset
		j = srcOffset
	)

	for i < dst.RequestNbSamples && j < src.LastNbSamples {
		dst.Data[dstCh][i] += src.Data[srcCh][j]
		i++
		j++
	}

	return j - srcOffset
}

func (src *Samples) Mix(dst *Samples, dstOffset int, srcOffset int) int {
	return src.MixChannels(dst, src.Format.Channels(), dstOffset, srcOffset)
}

// 相同声道之间混合
func (src *Samples) MixChannels(dst *Samples, srcChs []audio.Channel, dstOffset int, srcOffset int) int {
	if src.Format.Bits != audio.Bits_DEFAULT || dst == nil || dst.Format.Bits != audio.Bits_DEFAULT {
		return 0
	}
	var (
		i     int8 = 0
		j     int8 = 0
		mixed      = 0
	)

	for _, ch := range srcChs {
		if !ch.IsValid() {
			continue
		}

		i = src.ChannelIndex[ch]
		if i < 0 {
			continue
		}
		j = dst.ChannelIndex[ch]
		if j < 0 {
			continue
		}

		mixed = src.mixChannel(dst, j, i, dstOffset, srcOffset)
	}

	return mixed
}

func (src *Samples) CopyTo(dst *Samples, dstOffset, srcOffset int) int {
	if src.Format.Bits != audio.Bits_DEFAULT || dst == nil || dst.Format.Bits != audio.Bits_DEFAULT {
		return 0
	}
	var (
		i     int8 = 0
		j     int8 = 0
		mixed      = 0
	)
	for _, ch := range src.Format.Channels() {
		if !ch.IsValid() {
			continue
		}

		i = src.ChannelIndex[ch]
		if i < 0 {
			continue
		}
		j = dst.ChannelIndex[ch]
		if j < 0 {
			continue
		}
		mixed = copy(dst.Data[i][dstOffset:], src.Data[j][srcOffset:])
	}

	return mixed
}

// 从 planar 格式转换至 packed 格式
func (s *Samples) PackedBytes() []byte {
	var (
		p        = make([]byte, s.LastSamplesSize())
		chs      = int(s.Format.Layout.Count)
		bits     = s.Format.Sample.Bits.Size()
		i, b, ch int
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
		c          = 0
		dstTwoSize = dstSize
	)

	if dstOneSize > 1 {
		dstTwoSize = int(uintptr(dstS[1]) - uintptr(dstS[0]))
	}

	for c = 0; c < dstOneSize && c < int(s.Format.Layout.Count); c++ {
		dstCh = utils.MakeSlice[byte](dstS[c], dstTwoSize)
		copied = copy(dstCh, s.RawData[c])
	}

	return copied
}

// 复制C内存至GO内存。复制之前必须保证format一致
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
		c      = 0
	)

	srcOffset = s.Format.SamplesSize(srcOffset)
	srcTwoSize = s.Format.SamplesSize(srcTwoSize)
	dstOffset = s.Format.SamplesSize(dstOffset)
	dstTwoSIze := s.SamplesSize()

	if srcOffset > srcTwoSize {
		s.LastErr = fmt.Errorf("src offset %d large than buffer size %d", srcOffset, srcTwoSize)
		return 0
	}
	if dstOffset > dstTwoSIze {
		s.LastErr = fmt.Errorf("dst offset %d large than buffer size %d", dstOffset, s.SamplesSize())
		return 0
	}

	// C内存的样本数量可能不等于 s *Samples 的数量
	for c = 0; c < srcOneSize && c < int(s.Format.Count) && c < len(s.RawData); c++ {
		srcCh = utils.MakeSlice[byte](srcS[c], srcTwoSize)
		copied = copy(s.RawData[c][dstOffset:], srcCh[srcOffset:])
	}
	// if copied != srcTwoSize {
	// 	s.LastErr = fmt.Errorf("buffer size %d per channel not enough, need %d", s.SamplesSize(), srcTwoSize)
	// }
	if copied > (dstTwoSIze - dstOffset) {
		copied = dstTwoSIze - dstOffset
	}

	return copied
}

func sampleSize(duration time.Duration, format *audio.Format) int {
	if duration < time.Millisecond {
		// 默认取 10ms
		duration = 10 * time.Millisecond
	}
	return int(time.Duration(format.Sample.Rate.ToInt()) * duration / time.Second)
}

func NewSamplesDuration(duration time.Duration, format audio.Format) (s *Samples) {
	if !format.IsValid() {
		return nil
	}
	samples := sampleSize(duration, &format)
	s = newSamples(samples, format)
	s.autoSize = duration == 0
	return
}

func NewSamples(samples int, format audio.Format) (s *Samples) {
	if !format.IsValid() {
		return nil
	}
	s = newSamples(samples, format)
	return
}

func newSamples(samples int, format audio.Format) (s *Samples) {
	p := make([]byte, samples*format.Size())

	s = &Samples{
		Format: format,
	}
	reuseSamples(s, p, format)

	return
}

func NewFromBytes(p []byte, format audio.Format) (s *Samples) {
	// 向下取整
	samples := len(p) / format.Size()
	s = newSamples(samples, format)
	copy(s.buffer, p)
	s.LastNbSamples = samples
	return
}

func ReuseSamples(p []byte, format audio.Format) (s *Samples) {
	s = &Samples{
		Format: format,
	}
	reuseSamples(s, p, format)
	s.LastNbSamples = s.RequestNbSamples

	return s
}

func reuseSamples(s *Samples, p []byte, format audio.Format) {
	chs := int(format.Layout.Count)
	samples := len(p) / format.Size()

	s.buffer = p
	s.Data = make([][]float64, chs)
	s.RawData = make([][]byte, chs)
	s.RequestNbSamples = samples
	s.Format = format
	s.fmt = format
	s.ChannelIndex = *format.ChannelIndex()

	s.setLayout(format.Layout)
}

func (s *Samples) resizeSample(samples int, format audio.Format) {
	p := make([]byte, samples*format.Size())
	// copy(p, s.buffer)
	reuseSamples(s, p, format)
}

func (s *Samples) Resize(samples int, format audio.Format) {
	if !format.IsValid() || (format == s.Format && s.SamplesCount() == samples) {
		return
	}
	if (s.Format == format || !s.Format.LessThan(format)) && samples <= s.SamplesCount() {
		// 缓存空间足够大，变更格式即可
		s.RequestNbSamples = samples
		s.SetLayout(format.Layout)
		s.Format = format
		return
	}

	s.resizeSample(samples, format)
}

func (s *Samples) ResizeDuration(duration time.Duration, format audio.Format) {
	samples := sampleSize(duration, &format)
	s.Resize(samples, format)
}

func (s *Samples) ChannelSamples(ch audio.Channel) *Samples {
	si := s.ChannelIndex[ch]
	if si < 0 {
		return nil
	}

	format := s.Format
	format.Layout = audio.NewLayout(ch)

	ns := &Samples{}
	reuseSamples(ns, s.buffer, format)

	// 将任意的声道都映射至第一个
	for i := 0; i < int(audio.Channel_MAX); i++ {
		ns.ChannelIndex[i] = 0
	}

	ns.Data[0] = s.Data[si]
	ns.RawData[0] = s.RawData[si]

	return ns
}

func (s *Samples) WrapError(err error) {
	if err != nil {
		if s.LastErr != nil {
			s.LastErr = errors.Wrap(err, s.LastErr.Error())
		} else {
			s.LastErr = err
		}
	}
}
