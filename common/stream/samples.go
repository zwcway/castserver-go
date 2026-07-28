package stream

import (
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/utils"
)

type SampleConfig struct {
}

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

// BufferSize Buffer 总大小，包含所有声道
func (s *Samples) BufferSize() int {
	// return s.Format.AllSamplesSize(s.nbSamples)
	return len(s.buffer)
}

// AllSamplesCount 当前总样本数量
func (s *Samples) AllSamplesCount() int {
	return int(s.Format.Layout.Count) * s.LastNbSamples
}

// ChannelBytesSize 每声道中实际可存放的最大字节数
func (s *Samples) ChannelBytesSize() int {
	if len(s.buffer) == 0 {
		return 0
	}
	return len(s.buffer) / int(s.fmt.Count)
}

// SamplesCount 每声道中实际可存放的最大样本数
func (s *Samples) SamplesCount() int {
	return s.ChannelBytesSize() / s.fmt.Bits.Size()
}

// RequestSamplesSize 每声道中请求存放的最大字节数
func (s *Samples) RequestSamplesSize() int {
	return s.Format.SamplesSize(s.RequestNbSamples)
}

// LastSamplesSize 当前每声道已存入的字节数
func (s *Samples) LastSamplesSize() int {
	return s.Format.SamplesSize(s.LastNbSamples)
}

func (s *Samples) SetRate(f audio.Rate) {
	s.Format.Sample.Rate = f
}

// SetFormatAndIndex 在 SourceStreamer 中调用
func (s *Samples) SetFormatAndIndex(f audio.Format, ci audio.ChannelIndex) {
	if s.Format == f {
		return
	}
	s.Format = f
	s.ChannelIndex = ci
}

// SetLayout 更改声道布局
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
		chBuf     unsafe.Pointer
		perChSize = s.ChannelBytesSize()
		bits      = s.Format.Sample.Bits.Size()
		samples   = perChSize / 8
	)

	s.Format.Layout = f

	s.Data = make([][]float64, chs)
	s.RawData = make([][]byte, chs)

	for ch = range chs {
		chBuf = unsafe.Pointer(&s.buffer[ch*samples*bits])

		s.Data[ch] = utils.MakeSlice[float64](chBuf, samples)
		s.RawData[ch] = utils.MakeSlice[byte](chBuf, perChSize)
	}

	// s.setChannelIndex()
}

func (s *Samples) SetChannelIndex(ci audio.ChannelIndex) {
	if ci == nil {
		s.ChannelIndex = s.Format.ChannelIndex()
		return
	}

	s.ChannelIndex = ci
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
	clear(s.buffer)
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
		for k := j; k < len(ch); k++ {
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
	size := s.LastSamplesSize()
	return planar[:size]
}

func (s *Samples) ChannelIndexBytes(c int) []byte {
	planar := s.RawData[c]
	size := s.LastSamplesSize()
	return planar[:size]
}

// 混合声道，处理声道数量不一致的情况
func (s *Samples) MixChannelMap(dst *Samples, dstOffset, srcOffset int) (mixed int) {
	if sc := s.ChannelIndex[audio.ChannelFrontCenter]; sc >= 0 && s.Format.Layout.Count == 1 && dst.Format.Layout.Count > 1 {
		dc := dst.ChannelIndex[audio.ChannelFrontCenter]
		if dc >= 0 {
			mixed = s.mixChannel(dst, dc, sc, dstOffset, srcOffset)
			return
		}
		// 混合至前左和前右
		dc = dst.ChannelIndex[audio.ChannelFrontLeft]
		if dc >= 0 {
			mixed = s.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		dc = dst.ChannelIndex[audio.ChannelFrontRight]
		if dc >= 0 {
			mixed = s.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		return
	}

	if dc := dst.ChannelIndex[audio.ChannelFrontCenter]; dc >= 0 && s.Format.Layout.Count > 1 && dst.Format.Layout.Count == 1 {
		sc := s.ChannelIndex[audio.ChannelFrontCenter]

		if sc >= 0 {
			mixed = s.mixChannel(dst, dc, sc, dstOffset, srcOffset)
			return
		}
		// 从前左和前右混合
		sc = s.ChannelIndex[audio.ChannelFrontLeft]
		if sc >= 0 {
			mixed = s.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		sc = s.ChannelIndex[audio.ChannelFrontRight]
		if sc >= 0 {
			mixed = s.mixChannel(dst, dc, sc, dstOffset, srcOffset)
		}
		return
	}

	return s.MixChannels(dst, s.Format.Channels(), dstOffset, srcOffset)
}

func (s *Samples) mixChannel(dst *Samples, dstCh, srcCh int8, dstOffset, srcOffset int) int {
	var (
		i = dstOffset
		j = srcOffset
	)

	for i < dst.RequestNbSamples && j < s.LastNbSamples {
		dst.Data[dstCh][i] += s.Data[srcCh][j]
		i++
		j++
	}

	return j - srcOffset
}

func (s *Samples) Mix(dst *Samples, dstOffset int, srcOffset int) int {
	return s.MixChannels(dst, s.Format.Channels(), dstOffset, srcOffset)
}

// 相同声道之间混合
func (s *Samples) MixChannels(dst *Samples, srcChs []audio.Channel, dstOffset int, srcOffset int) int {
	if s.Format.Bits != audio.Bits_DEFAULT || dst == nil || dst.Format.Bits != audio.Bits_DEFAULT {
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

		i = s.ChannelIndex[ch]
		if i < 0 {
			continue
		}
		j = dst.ChannelIndex[ch]
		if j < 0 {
			continue
		}

		mixed = s.mixChannel(dst, j, i, dstOffset, srcOffset)
	}

	return mixed
}

func (s *Samples) CopyTo(dst *Samples, dstOffset, srcOffset int) int {
	if s.Format.Bits != audio.Bits_DEFAULT || dst == nil || dst.Format.Bits != audio.Bits_DEFAULT {
		return 0
	}
	var (
		i     int8 = 0
		j     int8 = 0
		mixed      = 0
	)
	for _, ch := range s.Format.Channels() {
		if !ch.IsValid() {
			continue
		}

		i = s.ChannelIndex[ch]
		if i < 0 {
			continue
		}
		j = dst.ChannelIndex[ch]
		if j < 0 {
			continue
		}
		mixed = copy(dst.Data[i][dstOffset:], s.Data[j][srcOffset:])
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

func sampleSize(duration time.Duration, format *audio.Format) int {
	if duration < time.Millisecond {
		// 默认取 10ms
		duration = 10 * time.Millisecond
	}
	return int(time.Duration(format.Sample.Rate.ToInt()) * duration / time.Second)
}

func NewSamplesDuration(duration time.Duration, format audio.Format) (s *Samples) {
	samples := sampleSize(duration, &format)
	s = newSamples(samples, format)
	s.autoSize = duration == 0
	return
}

func NewSamples(samples int, format audio.Format) (s *Samples) {
	s = newSamples(samples, format)
	return
}

func newSamples(samples int, format audio.Format) (s *Samples) {
	p := make([]byte, samples*format.Size())

	s = &Samples{
		Format: format,
	}

	s.reuseBuffer(p, format)

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

func NewReuse(p []byte, format audio.Format) (s *Samples) {
	s = &Samples{
		Format: format,
	}
	s.reuseBuffer(p, format)
	s.LastNbSamples = s.RequestNbSamples

	return s
}

func (s *Samples) reuseBuffer(p []byte, format audio.Format) {
	chs := int(format.Layout.Count)

	s.buffer = p
	s.Data = make([][]float64, chs)
	s.RawData = make([][]byte, chs)
	s.Format = format
	s.fmt = format
	s.ChannelIndex = format.ChannelIndex()

	if format.IsValid() {
		s.RequestNbSamples = len(p) / format.Size()
	}

	s.setLayout(format.Layout)
}

func (s *Samples) resize(samples int, format audio.Format) {
	p := make([]byte, samples*format.Size())
	// copy(p, s.buffer)
	s.reuseBuffer(p, format)
}

// Resize 按样本数量和格式重新调整缓存大小
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

	s.resize(samples, format)
}

// ResizeDuration 按时长调整缓存大小
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
	format.Layout = audio.NewLayouts(ch)

	ns := &Samples{}
	ns.reuseBuffer(s.buffer, format)

	// 将任意的声道都映射至第一个
	for i := 0; i < int(audio.ChannelMax); i++ {
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
