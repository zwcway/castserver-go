package websockets

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
)

type ResponseSpeakerInfo struct {
	ResponseSpeakerList

	Statistic speaker.Statistic `jp:"statisitc"`
}

func NewResponseSpeakerInfo(sp *speaker.Speaker) *ResponseSpeakerInfo {

	return &ResponseSpeakerInfo{
		ResponseSpeakerList: *NewResponseSpeakerList(sp),
		Statistic:           sp.Statistic,
	}
}

type ResponseSpeakerList struct {
	ID          int32             `jp:"id"`
	Name        string            `jp:"name"`
	IP          string            `jp:"ip"`
	MAC         string            `jp:"mac"`
	Channel     uint8             `jp:"ch,omitempty"`
	Line        *ResponseLineList `jp:"line,omitempty"`
	BitList     []string          `jp:"bitList,omitempty"`
	RateList    []int             `jp:"rateList,omitempty"`
	Bits        string            `jp:"bits"`
	Rate        int               `jp:"rate"`
	Volume      int               `jp:"vol,omitempty"`
	Mute        bool              `jp:"mute"`
	AbsoluteVol bool              `jp:"avol,omitempty"`
	PowerState  int               `jp:"power,omitempty"`
	ConnectTime int               `jp:"cTime,omitempty"`
}

func NewResponseSpeakerList(sp *speaker.Speaker) *ResponseSpeakerList {
	power := int(sp.PowerSate)
	if !sp.PowerSave {
		power = -1
	}
	ct := 0
	if !sp.ConnTime.IsZero() {
		ct = int(sp.ConnTime.Unix())
	}
	return &ResponseSpeakerList{
		ID:          int32(sp.Id),
		Name:        sp.Name,
		IP:          sp.IP.String(),
		MAC:         sp.MAC.String(),
		Channel:     uint8(sp.Channel),
		Line:        NewResponseLineList(sp.Line),
		BitList:     sp.BitsMask.Slice(),
		RateList:    sp.RateMask.Slice(),
		Rate:        sp.Rate.ToInt(),
		Bits:        sp.Bits.String(),
		Volume:      int(sp.Volume.Volume()),
		Mute:        sp.Volume.Mute(),
		AbsoluteVol: sp.AbsoluteVol,
		PowerState:  power,
		ConnectTime: ct,
	}
}

type ResponseLineSource struct {
	Rate     int    `jp:"rate"`
	Bits     string `jp:"bits"`
	Channels []int  `jp:"channels"`
	Layout   string `jp:"layout"`

	Type int `jp:"type"`

	// 文件播放
	Duration int `jp:"cur,omitempty"`
	Total    int `jp:"dur,omitempty"`
}

func NewResponseLineSource(line *speaker.Line) *ResponseLineSource {
	if line == nil {
		return nil
	}
	return &ResponseLineSource{
		Rate:     line.Input.Format.SampleRate.ToInt(),
		Bits:     line.Input.Format.SampleBits.String(),
		Channels: line.Input.Format.Layout.Mask.SliceInt(),
		Layout:   line.Input.Format.Layout.String(),
		Type:     int(line.Input.From),
		Duration: int(line.Input.Duration().Seconds()),
		Total:    int(line.Input.TotalDuration().Seconds() - 1),
	}
}

type ResponseLineList struct {
	ID     uint8  `jp:"id"`
	Name   string `jp:"name"`
	Volume int    `jp:"vol"`
	Mute   bool   `jp:"mute"`
}

func NewResponseLineList(ls *speaker.Line) *ResponseLineList {
	if ls == nil {
		return nil
	}
	return &ResponseLineList{
		ID:     uint8(ls.Id),
		Name:   ls.Name,
		Volume: int(ls.Volume.Volume() * 100),
		Mute:   ls.Volume.Mute(),
	}
}

type ResponseEqualizer struct {
	Switch     bool         `jp:"enable"`
	Equalizers [][3]float32 `jp:"eqs,omitempty"`
}

type ResponseLineInfo struct {
	ResponseLineList

	Channels   []int                  `jp:"chlist"`
	Layout     string                 `jp:"layout"`
	Speakers   []*ResponseSpeakerList `jp:"speakers,omitempty"`
	Input      *ResponseLineSource    `jp:"source,omitempty"`
	Equalizers *ResponseEqualizer     `jp:"eq,omitempty"`
}

func NewResponseEqualizer(line *speaker.Line) *ResponseEqualizer {
	if line == nil || line.Equalizer == nil {
		return nil
	}
	list := &ResponseEqualizer{
		Switch: line.Equalizer.IsOn(),
	}
	for _, e := range line.Equalizer.Equalizer() {
		list.Equalizers = append(list.Equalizers, [3]float32{float32(e.Frequency), float32(e.Gain), float32(e.Q)})
	}
	return list
}

func NewResponseLineInfo(line *speaker.Line) *ResponseLineInfo {
	if line == nil {
		return nil
	}
	info := &ResponseLineInfo{
		ResponseLineList: *NewResponseLineList(line),

		Channels:   line.Layout().Mask.SliceInt(),
		Layout:     line.Layout().String(),
		Speakers:   make([]*ResponseSpeakerList, line.SpeakerCount()),
		Input:      NewResponseLineSource(line),
		Equalizers: NewResponseEqualizer(line),
	}

	for i, s := range line.Speakers() {
		info.Speakers[i] = NewResponseSpeakerList(s)
	}

	return info
}

type ResponseChannelInfo struct {
	ID   uint8  `jp:"id"`
	Name string `jp:"name"`
}

func NewResponseChannelInfo(ch audio.Channel) *ResponseChannelInfo {
	if !ch.IsValid() {
		return nil
	}
	return &ResponseChannelInfo{
		ID:   uint8(ch),
		Name: ch.String(),
	}
}
