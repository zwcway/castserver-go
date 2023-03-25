package speaker

import (
	"errors"
	"fmt"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/utils"
)

var lineList []*Line = make([]*Line, 0)

type Line struct {
	Id   LineID
	Name string
	UUID string // dlna 标识

	chRoute  []audio.ChannelRoute // 输出的声道路由关系表
	speakers []*Speaker

	Input  stream.Source // 输入格式
	Output audio.Format  // 输出格式

	Mixer     stream.MixerElement
	Volume    stream.VolumeElement
	Spectrum  stream.SpectrumElement
	Equalizer stream.EqualizerElement
	Player    stream.RawPlayerElement
	Resample  stream.ResampleElement
	Pusher    stream.SwitchElement

	isDeleted bool
}

func (l *Line) Layout() *audio.ChannelLayout {
	return &l.Output.Layout
}

// 如果返回值不空，就表示有speaker
func (l *Line) Channels() []audio.Channel {
	return l.Output.Layout.Mask.Slice()
}

func (l *Line) ChannelRoute(dst audio.Channel) []audio.Channel {
	for _, cr := range l.chRoute {
		if cr.To == dst {
			return cr.From
		}
	}
	return []audio.Channel{dst}
}

func (l *Line) Speakers() []*Speaker {
	return l.speakers
}

func (l *Line) SpeakerCount() int {
	return len(l.speakers)
}

func (l *Line) SpeakersByChannel(ch audio.Channel) []*Speaker {
	sps := make([]*Speaker, 0)

	for i := 0; i < len(l.speakers); i++ {
		if l.speakers[i].Channel == ch {
			sps = append(sps, l.speakers[i])
		}
	}
	return sps
}

func (l *Line) AppendSpeaker(sp *Speaker) {
	l.speakers = append(l.speakers, sp)
	l.refresh()
	bus.Trigger("line speaker appended", l, sp)
}

func (line *Line) RemoveSpeakerById(spid ID) {
	for i, sp := range line.speakers {
		if sp.Id == spid {
			utils.SliceQuickRemove(&line.speakers, i)
			bus.Trigger("line speaker removed", line, sp)
			line.refresh()
			break
		}
	}
}

func (l *Line) RemoveSpeaker(sp *Speaker) {
	l.RemoveSpeakerById(sp.Id)
}

func (l *Line) SetOutput(f audio.Format) {
	if l.Output.Equal(&f) {
		return
	}

	l.Output = f
	l.refresh()

	bus.Trigger("line output changed", l, f)
}

func (l *Line) refresh() {
	channels := []audio.Channel{}
	for i := 0; i < len(l.speakers); i++ {
		channels = append(channels, (l.speakers[i].Channel))
	}
	l.Output.Layout = audio.NewChannelLayout(channels...)

	bus.Trigger("line refresh", l)
}

func (l *Line) Elements() []stream.Element {
	return []stream.Element{
		l.Mixer,
		l.Equalizer,
		l.Player,
		l.Spectrum,
		l.Volume,
		l.Pusher,
	}
}

func (l *Line) IsDeleted() bool {
	return l.isDeleted
}

func (l *Line) ApplyInput(ss stream.SourceStreamer) {
	l.Input.ApplySource(ss)
	bus.Trigger("line input changed", l, ss)

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
	l.Input.Format = inFormat
	if fs, ok := ss.(stream.FileStreamer); ok {
		format := audio.InternalFormat()
		format.InitFrom(inFormat)
		fs.SetOutFormat(format)
	}
	bus.Trigger("line audiofile opened", l, ss, inFormat, outFormat)
}

func LineList() []*Line {
	return lineList
}

func FindLineByID(id LineID) *Line {
	for _, l := range lineList {
		if l.Id == id {
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
	if m > maxLineID && m < LineID_MAX {
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
		src.speakers[i].Line = dst
	}
	dst.refresh()

	removeLine(id)

	src.isDeleted = true

	bus.Trigger("line deleted", src, dst)

	return nil
}

func CountLine() int {
	return len(lineList)
}

func removeLine(id LineID) {
	l := len(lineList) - 1
	for i := 0; i <= l; i++ {
		if lineList[i].Id == id {
			if i != l {
				lineList[i] = lineList[l]
			}
			lineList = lineList[:l]
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

func initLine() error {
	NewLine("Default")

	return nil
}

func NewLine(name string) *Line {
	locker.Lock()
	defer locker.Unlock()

	var line Line

	l := len(lineList)
	if l >= int(LineID_MAX) {
		return nil
	}

	line.Id = getLineID()
	line.Name = name
	line.UUID = generateUUID(name)

	line.SetOutput(audio.DefaultFormat())

	line.Player = element.NewPlayer()
	line.Volume = element.NewVolume(0.1)
	line.Mixer = element.NewEmptyMixer()
	line.Spectrum = element.NewSpectrum()
	line.Equalizer = element.NewEqualizer(nil)
	line.Resample = element.NewResample(line.Output)

	line.Input.Mixer = line.Mixer
	line.Input.PipeLine = pipeline.NewPipeLine(line.Output,
		line.Mixer,
		line.Equalizer,
		line.Player,
		line.Spectrum,
		line.Volume,
		//line.Pusher, Resample 放到 pusher 中处理，否则声道路由功能不方便实现
	)

	lineList = append(lineList, &line)

	bus.Trigger("line created", &line)

	return &line
}
