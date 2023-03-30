package speaker

import (
	"errors"
	"fmt"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
)

var lineList []*Line = make([]*Line, 0)

type Line struct {
	ID   LineID `gorm:"primaryKey;column:id"`
	Name string `gorm:"column:name"`
	UUID string `gorm:"column:uuid"` // dlna 标识

	Volume uint8 `gorm:"column:volume"`
	Mute   bool  `gorm:"column:mute"`

	EQ      DBeqData       `gorm:"column:eq"`    // 均衡器
	ChRoute DBChannelRoute `gorm:"column:route"` // 输出的声道路由关系表

	CreatedAt time.Time
	UpdatedAt time.Time

	Input  stream.Source `gorm:"-"` // 输入格式
	Output audio.Format  `gorm:"-"` // 输出格式

	MixerEle     stream.MixerElement     `gorm:"-"`
	VolumeEle    stream.VolumeElement    `gorm:"-"`
	SpectrumEle  stream.SpectrumElement  `gorm:"-"`
	EqualizerEle stream.EqualizerElement `gorm:"-"`
	PlayerEle    stream.RawPlayerElement `gorm:"-"`
	// ResampleEle  stream.ResampleElement  `gorm:"-"`
	// PusherEle    stream.SwitchElement    `gorm:"-"`

	isDeleted bool

	speakers []*Speaker   `gorm:"-"`
	spsByCh  [][]*Speaker `gorm:"-"`
}

func (l *Line) String() string {
	return fmt.Sprintf("%s(%d)", l.Name, l.ID)
}

func (l *Line) Layout() audio.ChannelLayout {
	return l.Output.Layout
}

// 如果返回值不空，就表示有speaker
func (l *Line) Channels() []audio.Channel {
	return l.Output.Layout.Mask.Slice()
}

func (l *Line) ChannelRoute(dst audio.Channel) []audio.Channel {
	for _, cr := range l.ChRoute.R {
		if cr.To == dst {
			return cr.From
		}
	}
	return []audio.Channel{dst}
}

func (l *Line) AddRoute(to audio.Channel, from audio.Channel) {
	for i, cr := range l.ChRoute.R {
		if cr.To != to {
			continue
		}
		l.ChRoute.R[i].From = append(l.ChRoute.R[i].From, from)
		l.changeRoute()
		return
	}

	l.ChRoute.R = append(l.ChRoute.R, audio.ChannelRoute{
		From: []audio.Channel{from},
		To:   to,
	})
	l.changeRoute()
}

func (l *Line) RemoveRoute(to audio.Channel, from audio.Channel) {
	for i, cr := range l.ChRoute.R {
		if cr.To != to {
			continue
		}
		if from.IsValid() {
			utils.SliceQuickRemoveItem(&l.ChRoute.R[i].From, from)
		} else {
			utils.SliceQuickRemove(&l.ChRoute.R, i)
		}
		l.changeRoute()
		return
	}
}

func (l *Line) changeRoute() {
	bus.Dispatch("line edited", l, "route", l.ChRoute)
}

func (l *Line) syncRoute() {
	bus.Dispatch("line route changed", l)
}

func (l *Line) Equalizer() *dsp.DataProcess {
	return l.EQ.Eq
}

func (l *Line) SetEqualizer(eq *dsp.DataProcess) (err error) {
	l.EQ.Eq = eq

	l.syncEqualizer()

	bus.Dispatch("line edited", l, "eq", l.EQ)

	return nil
}

func (l *Line) syncEqualizer() {
	if l.EQ.Eq == nil {
		l.EQ.Eq = dsp.NewDataProcess(0)
	}
	eq := l.EQ.Eq
	l.EqualizerEle.SetDelay(eq.Delay)
	l.EqualizerEle.SetFilterType(eq.Type)
	l.EqualizerEle.SetEqualizer(eq.FEQ)

	bus.Dispatch("line eq changed", l)
}

func (l *Line) Speakers() []*Speaker {
	return l.speakers
}

func (l *Line) SpeakerCount() int {
	return len(l.speakers)
}

func (l *Line) SpeakersByChannel(ch audio.Channel) []*Speaker {
	return l.spsByCh[ch]
}

func (l *Line) AppendSpeaker(sp *Speaker) {
	if utils.SliceContains(l.speakers, sp) < 0 {
		l.speakers = append(l.speakers, sp)
	}

	l.refresh()
	bus.Dispatch("line speaker appended", l, sp)
}

func (line *Line) RemoveSpeakerById(spid SpeakerID) {
	for i, sp := range line.speakers {
		if sp.ID == spid {
			utils.SliceQuickRemove(&line.speakers, i)
			bus.Dispatch("line speaker removed", line, sp)
			line.refresh()
			break
		}
	}
}

func (l *Line) RemoveSpeaker(sp *Speaker) {
	l.RemoveSpeakerById(sp.ID)
}

func (l *Line) SetOutput(f audio.Format) {
	if l.Output.Equal(&f) {
		return
	}

	l.Output = f
	for _, sp := range l.speakers {
		sp.SetFormat(f)
	}

	bus.Dispatch("line output changed", l)
}

func (l *Line) SetVolume(vol uint8, mute bool) {
	l.Volume = vol
	l.Mute = mute
	l.Save()

	l.VolumeEle.SetVolume(float64(vol) / 100)
	l.VolumeEle.SetMute(mute)

	bus.Dispatch("line volume changed", l)
}

func (l *Line) SetName(n string) {
	l.Name = n
	bus.Dispatch("line edited", l, "name", n)
	bus.Dispatch("line name changed", l)
}

func (l *Line) refresh() {
	channels := []audio.Channel{}

	for i := 0; i < len(l.spsByCh); i++ {
		l.spsByCh[i] = nil
	}

	for _, sp := range l.speakers {
		l.spsByCh[sp.Channel] = append(l.spsByCh[sp.Channel], sp)
		channels = append(channels, sp.SampleChannel())
	}

	format := audio.Format{}
	format.SampleRate, format.SampleBits = l.decideOutputFormat()
	format.Layout = audio.NewChannelLayout(channels...)
	l.SetOutput(format)

	bus.Dispatch("line refresh", l)
}

func (l *Line) SpeakerSamplesFromat() (rm audio.RateMask, bm audio.BitsMask) {
	rm.CombineSlice(config.SupportAudioRates)
	bm.CombineSlice(config.SupportAudioBits)
	for _, sp := range l.speakers {
		rm.Intersect(sp.Config.RateMask)
		bm.Intersect(sp.Config.BitsMask)
	}
	return
}

func (l *Line) decideOutputFormat() (r audio.Rate, b audio.Bits) {
	rm, bm := l.SpeakerSamplesFromat()
	r = rm.Max()
	b = bm.Max()

	if rm.Isset(l.Input.Format.SampleRate) {
		r = l.Input.Format.SampleRate
	}
	if bm.Isset(l.Input.Format.SampleBits) {
		b = l.Input.Format.SampleBits
	}

	return
}

func (l *Line) Elements() []stream.Element {
	return []stream.Element{
		l.MixerEle,
		l.EqualizerEle,
		l.PlayerEle,
		l.SpectrumEle,
		l.VolumeEle,
	}
}

func (l *Line) IsDeleted() bool {
	return l.isDeleted
}

func (l *Line) ApplyInput(ss stream.SourceStreamer) {
	l.Input.ApplySource(ss)
	bus.Dispatch("line input changed", l, ss)

	bus.Register("audiofile opened", func(a ...any) error {
		tfs := a[0].(stream.FileStreamer)
		if tfs != ss {
			return nil
		}
		inFormat := a[1].(audio.Format)
		outFormat := a[2].(audio.Format)
		l.onInputChanged(ss, inFormat, outFormat)
		return nil
	})
}

func (l *Line) onInputChanged(ss stream.SourceStreamer, inFormat, outFormat audio.Format) {
	if fs, ok := ss.(stream.FileStreamer); ok {
		format := audio.InternalFormat()
		format.InitFrom(inFormat)
		fs.SetOutFormat(format)
	}
	if !l.Input.Format.Equal(&inFormat) {
		l.Input.Format = inFormat
		l.refresh()
	}
	bus.Dispatch("line audiofile opened", l, ss, inFormat, outFormat)
}

func (l *Line) Save() {
	bus.Dispatch("save line", l)
}

func (line *Line) init() {
	line.spsByCh = make([][]*Speaker, audio.Channel_MAX)

	line.SetOutput(audio.DefaultFormat())

	line.PlayerEle = element.NewPlayer()
	line.VolumeEle = element.NewVolume(float64(line.Volume) / 100)
	line.MixerEle = element.NewEmptyMixer()
	line.SpectrumEle = element.NewSpectrum()
	line.EqualizerEle = element.NewEqualizer(line.EQ.Eq)
	// line.ResampleEle = element.NewResample(line.Output)

	line.Input.Mixer = line.MixerEle
	line.Input.PipeLine = pipeline.NewPipeLine(line.Output,
		line.MixerEle,
		line.EqualizerEle,
		line.PlayerEle,
		line.SpectrumEle,
		line.VolumeEle,
		//line.Pusher, Resample 放到 pusher 中处理，否则声道路由功能不方便实现
	)

	line.syncEqualizer()
	line.syncRoute()

}

func LineList() []*Line {
	return lineList
}

func FindLineByID(id LineID) *Line {
	for _, l := range lineList {
		if l.ID == id {
			return l
		}
	}

	return nil
}

func getLineID() (m LineID) {
	m = maxLineID

	for FindLineByID(m) != nil {
		m++
		if m > LineID_MAX {
			m = DefaultLineID + 1
		}
	}
	if m >= maxLineID && m < LineID_MAX {
		maxLineID = m + 1
	}

	return m
}

func FindLineByUUID(uuid string) *Line {
	for i := 0; i < len(lineList); i++ {
		if lineList[i].UUID == uuid {
			return lineList[i]
		}
	}

	return nil
}

func DelLine(id LineID, move LineID) error {
	locker.Lock()
	defer locker.Unlock()

	if id == move {
		return nil
	}

	if id == DefaultLineID {
		return errors.New("can not delete default line")
	}

	src := FindLineByID(id)
	if src == nil {
		return &UnknownLineError{id}
	}

	dst := FindLineByID(move)
	if dst == nil {
		return &UnknownLineError{move}
	}

	// 迁移至新的线路
	for i := 0; i < len(src.speakers); i++ {
		sp := src.speakers[i]
		sp.LineId = dst.ID
		sp.Save()
		sp.Line = dst
	}
	dst.refresh()

	removeLine(id)

	src.isDeleted = true

	bus.Dispatch("line deleted", src, dst)

	return nil
}

func CountLine() int {
	return len(lineList)
}

func removeLine(id LineID) {
	l := len(lineList) - 1
	for i := 0; i <= l; i++ {
		if lineList[i].ID == id {
			if i != l {
				lineList[i] = lineList[l]
			}
			lineList = lineList[:l]
			return
		}
	}
}

func FindSpeakersByLine(line LineID) []*Speaker {
	l := FindLineByID(line)
	if l == nil {
		return nil
	}

	return l.Speakers()
}

func FindSpeakersByChannel(line LineID, ch audio.Channel) []*Speaker {
	l := FindLineByID(line)
	if l == nil {
		return nil
	}

	return l.SpeakersByChannel(ch)
}

func DefaultLine() *Line {
	return FindLineByID(DefaultLineID)
}

func generateUUID(name string) string {
	for {
		uuid := utils.MakeUUID(fmt.Sprintf("%s%d", name, time.Now().Nanosecond()))
		l := FindLineByUUID(uuid)
		if l == nil {
			return uuid
		}
	}
}

func NewLine(name string) *Line {
	locker.Lock()
	defer locker.Unlock()

	var line Line

	l := len(lineList)
	if l >= int(LineID_MAX) {
		return nil
	}

	line.ID = getLineID()
	line.Name = name
	line.UUID = generateUUID(name)
	line.Volume = 50
	line.init()
	lineList = append(lineList, &line)

	line.Save()

	bus.Dispatch("line created", &line)

	return &line
}
