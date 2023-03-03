package speaker

import (
	"errors"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/dsp"
)

type Input struct {
}
type Line struct {
	ID   LineID
	Name string
	UUID string // dlna 标识

	channels audio.ChannelMask

	Input  *audio.Format // 输入格式
	Output *audio.Format // 输出格式

	Volume     int
	Spectrum   []float32
	LevelMeter float64

	Equalizer []dsp.FreqEqualizer
}

func (l *Line) SetVolume(vol int) {
	l.Volume = vol
}

func (l *Line) Channels() audio.ChannelMask {
	return l.channels
}

var lineList map[LineID]*Line
var listByLine speakerMapSlice

var DefaultLineID LineID = 0

func initLine() error {
	listByLine = make(speakerMapSlice, 0)
	lineList = make(map[LineID]*Line, 1)
	AddLine("Default")

	return nil
}

func lineIsValid(line LineID) bool {
	_, ok := lineList[line]
	return ok
}

func AddLine(name string) *Line {
	var line Line

	l := len(lineList)
	if l >= 255 {
		return nil
	}
	line.ID = DefaultLineID

	if l > 0 {
		for i := DefaultLineID; i < 255; i++ {
			if _, ok := lineList[i]; !ok {
				line.ID = i
				break
			}
		}
	}

	line.Name = name

	lineList[line.ID] = &line

	return &line
}

func LineList() map[LineID]*Line {
	return lineList
}

func FindLineByID(id LineID) *Line {
	if line, ok := lineList[id]; ok {
		return line
	}
	return nil
}

func EditLine(id LineID, name string) error {
	line, ok := lineList[id]
	if !ok {
		return &UnknownLineError{id}
	}
	line.Name = name

	return nil
}

func DelLine(id LineID, move LineID) error {
	lock.Lock()
	defer lock.Unlock()

	if id == move {
		return nil
	}

	if id == DefaultLineID {
		return errors.New("can not delete default line")
	}

	if !lineIsValid(move) {
		return &UnknownLineError{move}
	}

	// 迁移至新的线路
	for _, sp := range list {
		if sp.Line == id {
			sp.Line = move
		}
	}

	delete(lineList, id)

	return nil
}

func CountLine() int {
	return len(lineList)
}

func CountLineSpeaker(id LineID) int {
	return listByLine.len(int(id))
}

func appendSpeakerToLine(sp *Speaker) {
	listByLine.add(int(sp.Line), sp)
	refreshLine(sp.Line)
}

func removeSpeakerFromLine(sp *Speaker) {
	listByLine.remove(int(sp.Line), sp)
	refreshLine(sp.Line)
}

func refreshLine(line LineID) {
	l, ok := lineList[line]
	if !ok {
		return
	}

	channels := []uint8{}
	for _, sp := range FindSpeakersByLine(line) {
		channels = append(channels, uint8(sp.Channel))
	}
	l.channels, _ = audio.NewAudioChannelMask(channels)
}

func FindSpeakersByLine(line LineID) []*Speaker {
	l, ok := listByLine[int(line)]
	if !ok {
		return nil
	}
	return l
}

func FindSpeakersByChannel(line LineID, ch audio.Channel) []*Speaker {
	sps := make([]*Speaker, 0)

	for _, sp := range FindSpeakersByLine(line) {
		if sp.Channel == ch {
			sps = append(sps, sp)
		}
	}
	return sps
}

func SetLineInput(line LineID, format *audio.Format) {
	l := FindLineByID(line)
	if l == nil {
		return
	}
	l.Input = format
}

func DefaultLine() *Line {
	return FindLineByID(DefaultLineID)
}
