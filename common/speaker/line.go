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
	ID       LineID `gorm:"primaryKey;column:id"`
	LineName string `gorm:"column:name"`
	UUID     string `gorm:"column:uuid"` // dlna 标识

	Volume uint8 `gorm:"column:volume"`
	Mute   bool  `gorm:"column:mute"`

	EQ      DBeqData       `gorm:"column:eq"`    // 均衡器
	ChRoute DBChannelRoute `gorm:"column:route"` // 输出的声道路由关系表

	CreatedAt time.Time
	UpdatedAt time.Time

	Input  stream.Source `gorm:"-"` // 输入格式
	Output audio.Format  `gorm:"-"` // 输出格式

	isDeleted bool

	speakers []*Speaker   `gorm:"-"`
	spsByCh  [][]*Speaker `gorm:"-"`
}

func (l *Line) String() string {
	return fmt.Sprintf("%s(%d)", l.LineName, l.ID)
}

func (l *Line) Name() string {
	return l.LineName
}

func (l *Line) Layout() audio.Layout {
	return l.Output.Layout
}

// 如果返回值不空，就表示有speaker
func (l *Line) Channels() []audio.Channel {
	return l.Output.Channels()
}

// 多个源声道路由至同一个目的声道
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
	l.Dispatch("line edited", "route", l.ChRoute)
}

func (l *Line) syncRoute() {
	l.Dispatch("line route changed")
}

func (l *Line) Equalizer() *dsp.EqualizerProcessor {
	return l.EQ.Eq
}

func (l *Line) SetEqualizer(eq *dsp.EqualizerProcessor) (err error) {
	l.EQ.Eq = eq

	l.syncEqualizer()

	l.Dispatch("line edited", "eq", l.EQ)

	return nil
}

func (l *Line) syncEqualizer() {
	if l.EQ.Eq == nil {
		l.EQ.Eq = dsp.NewPeakingFilterEqualizerProcessor(0)
	}
	eq := l.EQ.Eq
	l.Input.EqualizerEle.SetDelay(eq.Delay)
	l.Input.EqualizerEle.SetFilterType(eq.Type)
	l.Input.EqualizerEle.SetEqualizer(eq.Filters)

	l.Dispatch("line eq changed")
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

	l.SetOutput(l.decideOutputFormat())

	BusLineSpeakerAppended.Dispatch(l, sp)
	l.refresh()
}

func (line *Line) RemoveSpeakerById(spid SpeakerID) {
	for i, sp := range line.speakers {
		if sp.ID == spid {
			utils.SliceQuickRemove(&line.speakers, i)
			BusLineSpeakerRemoved.Dispatch(line, sp)
			line.refresh()
			break
		}
	}
}

func (l *Line) RemoveSpeaker(sp *Speaker) {
	l.RemoveSpeakerById(sp.ID)
}

// 更改输出格式
func (l *Line) SetOutput(f audio.Format) {
	if l.Output.Equal(f) {
		return
	}

	old := l.Output
	l.Output = f

	for _, sp := range l.speakers {
		sp.SetSample(f.Sample)
	}

	BusLineOutputChanged.Dispatch(l, &old)
}

func (l *Line) SetVolume(vol uint8, mute bool) {
	old := l.Input.VolumeEle.Volume()

	args := []any{}
	if l.Volume != vol {
		args = append(args, "volume", vol)
		l.Volume = vol
	}
	if l.Mute != mute {
		args = append(args, "mute", mute)
		l.Mute = mute
	}
	BusLineEdited.Dispatch(l, args...)

	l.Input.VolumeEle.SetVolume(float64(vol) / 100)
	l.Input.VolumeEle.SetMute(mute)

	BusLineVolumeChanged.Dispatch(l, old)
}

func (l *Line) SetName(n string) {
	old := l.LineName
	l.LineName = n
	BusLineEdited.Dispatch(l, "name", n)
	BusLineNameChanged.Dispatch(l, &old)
}

func (l *Line) refresh() {
	for i := 0; i < len(l.spsByCh); i++ {
		l.spsByCh[i] = make([]*Speaker, 0)
	}

	for _, sp := range l.speakers {
		l.spsByCh[sp.Channel] = append(l.spsByCh[sp.Channel], sp)
	}

	BusLineRefresh.Dispatch(l)
}

// 计算设备的格式支持情况
func (l *Line) SpeakerSamplesFromat() (rm audio.RateMask, bm audio.BitsMask) {
	rm.CombineSlice(config.SupportAudioRates)
	bm.CombineSlice(config.SupportAudioBits)
	for _, sp := range l.speakers {
		rm.Intersect(sp.Config.RateMask)
		bm.Intersect(sp.Config.BitsMask)
	}
	return
}

// 根据设备的支持程度，自动确定输出格式
func (l *Line) decideOutputFormat() audio.Format {
	channels := []audio.Channel{}
	format := audio.DefaultFormat()

	rm, bm := l.SpeakerSamplesFromat()
	if rm.IsValid() && bm.IsValid() {
		format.Rate = rm.Max()
		format.Bits = bm.Max()
	}

	// if rm.Isset(l.Output.Rate) {
	// 	format.Rate = l.Output.Rate
	// }
	// if bm.Isset(l.Output.Bits) {
	// 	format.Bits = l.Output.Bits
	// }

	for _, sp := range l.speakers {
		channels = append(channels, sp.SampleChannel())
	}
	format.Layout = audio.NewLayout(channels...)

	return format
}

func (l *Line) IsDeleted() bool {
	return l.isDeleted
}

func (l *Line) ApplyInput(ss stream.SourceStreamer) {
	l.Input.ApplySource(ss)
	BusLineInputChanged.Dispatch(l, ss)
}

func (l *Line) Save() {
	l.Dispatch("save line")
}

func (line *Line) init() {
	line.spsByCh = make([][]*Speaker, audio.Channel_MAX)

	line.Input.MixerEle = element.NewMixer()
	line.Input.VolumeEle = element.NewVolume(float64(line.Volume) / 100)
	line.Input.SpectrumEle = element.NewSpectrum()
	line.Input.EqualizerEle = element.NewEqualizer(line.EQ.Eq)
	line.Input.PlayerEle = element.NewPlayer()

	line.Input.PipeLine = pipeline.NewPipeLine(line.Output,
		line.Input.MixerEle,
		line.Input.EqualizerEle,
		line.Input.PlayerEle,
		line.Input.SpectrumEle,
		line.Input.VolumeEle,
	)

	line.syncEqualizer()
	line.syncRoute()

	line.SetOutput(line.decideOutputFormat())
}

func (o *Line) Dispatch(e string, args ...any) error {
	return bus.DispatchObj(o, e, args...)
}

func (o *Line) Register(e string, c func(o any, a ...any) error) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
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

	BusLineDeleted.Dispatch(src, dst)

	src.Input.PipeLine.Close()
	bus.UnregisterObj(src)

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
	line.LineName = name
	line.UUID = generateUUID(name)
	line.Volume = 50
	line.init()
	lineList = append(lineList, &line)

	line.Save()

	BusLineCreated.Dispatch(&line)

	return &line
}
