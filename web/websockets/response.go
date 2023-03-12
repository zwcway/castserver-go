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
		Bits:        sp.Bits.Name(),
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
	Channels int    `jp:"channels"`

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
		Bits:     line.Input.Format.SampleBits.Name(),
		Channels: line.Input.Format.Layout.Count,
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

type ResponseLineInfo struct {
	ResponseLineList

	Speakers   []*ResponseSpeakerList `jp:"speakers,omitempty"`
	Input      *ResponseLineSource    `jp:"source,omitempty"`
	Equalizers [][3]float32           `jp:"eq,omitempty"`
	EqSwitch   bool                   `jp:"eqenable"`
}

func NewResponseEqualizer(line *speaker.Line) [][3]float32 {
	if line == nil || line.Equalizer == nil {
		return nil
	}
	list := [][3]float32{}
	for _, e := range line.Equalizer.Equalizer() {
		list = append(list, [3]float32{float32(e.Frequency), float32(e.Gain), float32(e.Q)})
	}
	return list
}

func NewResponseLineInfo(line *speaker.Line) *ResponseLineInfo {
	if line == nil {
		return nil
	}
	info := &ResponseLineInfo{
		ResponseLineList: *NewResponseLineList(line),

		Speakers:   make([]*ResponseSpeakerList, line.SpeakerCount()),
		Input:      NewResponseLineSource(line),
		Equalizers: NewResponseEqualizer(line),
		EqSwitch:   line.Equalizer.IsOn(),
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
		Name: ch.Name(),
	}
}
