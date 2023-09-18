package speaker

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/stream"
)

var speakerList []*Speaker = make([]*Speaker, 0)

type SpeakerConfig struct {
	ID          uint           `gorm:"column:id;primaryKey"`
	RateMask    audio.RateMask `gorm:"culumn:rate_mask"`    // 设备支持的采样率列表
	BitsMask    audio.BitsMask `gorm:"culumn:bits_mask"`    // 设备支持的位宽列表
	AbsoluteVol bool           `gorm:"culumn:absolute_vol"` // 支持绝对音量控制
	PowerSave   bool           `gorm:"culumn:power_save"`   // 是否支持电源控制
}

type Speaker struct {
	ID          SpeakerID `gorm:"column:id;primaryKey"`
	LineId      LineID    `gorm:"column:line_id;index"`
	SpeakerName string    `gorm:"column:name"`
	Supported   bool      `gorm:"column:supported"` // 是否兼容

	Mac   string `gorm:"column:mac"`
	Ip    string `gorm:"column:ip"`
	Mode  Model  `gorm:"column:mode"`
	Dport uint16 `gorm:"column:dport"` // pcm data port

	Rate    uint8  `gorm:"column:rate"`
	Bits    uint8  `gorm:"column:bits"`
	Channel uint32 `gorm:"column:channel"` // 当前设置的声道

	PowerState PowerState `gorm:"column:power_state"` // 当前电源状态

	Volume uint8 `gorm:"column:volume"`
	Mute   bool  `gorm:"column:mute"`

	Config SpeakerConfig `gorm:"foreignKey:ID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	CreatedAt time.Time
	UpdatedAt time.Time

	/******************************* Model end ******************************************/

	State State `gorm:"-"` // 当前连接状态

	Line *Line `gorm:"-"`

	PipeLine     stream.PipeLiner        `gorm:"-"`
	MixerEle     stream.MixerElement     `gorm:"-"`
	VolumeEle    stream.VolumeElement    `gorm:"-"` // 音量
	SpectrumEle  stream.SpectrumElement  `gorm:"-"`
	EqualizerEle stream.EqualizerElement `gorm:"-"`
	PlayerEle    stream.RawPlayerElement `gorm:"-"`

	ConnTime time.Time      `gorm:"-"`
	Conn     *net.UDPConn   `gorm:"-"`
	Queue    chan QueueData `gorm:"-"`

	Timeout   int       `gorm:"-"` // 超时计数
	Statistic Statistic `gorm:"-"`

	isDeleted bool
}

func (sp *Speaker) String() string {
	return fmt.Sprintf("%s(%s/%s)", sp.SpeakerName, sp.Ip, sp.Mac)
}

func (sp *Speaker) Name() string {
	return sp.SpeakerName
}

func (sp *Speaker) SetName(n string) {
	sp.SpeakerName = n
	bus.DispatchObj(sp, "speaker edited", "name", n)
	bus.DispatchObj(sp, "speaker name changed")
}

func (sp *Speaker) SetIP(ip netip.Addr) {
	sp.Ip = ip.String()
}

func (sp *Speaker) Format() audio.Format {
	return audio.Format{
		Sample: audio.Sample{
			Rate: sp.SampleRate(),
			Bits: sp.SampleBits(),
		},
		Layout: sp.Layout(),
	}
}
func (sp *Speaker) SampleRate() audio.Rate {
	return audio.Rate(sp.Rate)
}

func (sp *Speaker) SampleBits() audio.Bits {
	return audio.Bits(sp.Bits)
}

func (sp *Speaker) Layout() audio.Layout {
	return audio.NewLayout(sp.SampleChannel())
}

func (sp *Speaker) SetSample(f audio.Sample) {
	if sp.Rate == uint8(f.Rate) && sp.Bits == uint8(f.Bits) {
		return
	}

	sp.Rate = uint8(f.Rate)
	sp.Bits = uint8(f.Bits)
	bus.DispatchObj(sp, "speaker edited", "rate", sp.Rate, "bits", sp.Bits)
	bus.DispatchObj(sp, "speaker format changed")
}

func (sp *Speaker) SampleChannel() audio.Channel {
	return audio.Channel(sp.Channel)
}

func (sp *Speaker) SetChannel(ch audio.Channel) {
	if ch.IsValid() {
		sp.Channel = uint32(ch)
	} else {
		sp.Channel = uint32(audio.Channel_NONE)
	}
	if sp.Line != nil {
		sp.Line.refresh()
	}
	bus.DispatchObj(sp, "speaker edited", "channel", sp.Channel)
	bus.DispatchObj(sp, "speaker channel changed")
}

func (sp *Speaker) IsOnline() bool {
	return sp.State == State_ONLINE
}
func (sp *Speaker) IsOffline() bool {
	return sp.State == State_OFFLINE
}
func (sp *Speaker) IsSupported() bool {
	return sp.Supported
}

func (sp *Speaker) CheckOnline() {
	if sp.Dport > 0 {
		sp.SetOnline()
	} else {
		sp.SetOffline()
	}
}

func (sp *Speaker) SetVolume(vol uint8, mute bool) {
	sp.Volume = vol
	sp.Mute = mute

	sp.VolumeEle.SetVolume(float64(vol) / 100)
	sp.VolumeEle.SetMute(mute)

	bus.DispatchObj(sp, "speaker edited", "volume", vol, "mute", mute)
	bus.DispatchObj(sp, "speaker volume changed")
}

func (sp *Speaker) SetOffline() {
	sp.State &= ^State_ONLINE
}

func (sp *Speaker) SetOnline() {
	sp.State |= State_ONLINE
}

func (sp *Speaker) UDPAddr() *net.UDPAddr {
	addr, _ := net.ResolveUDPAddr("udp", sp.Ip+":"+strconv.Itoa(int(sp.Dport)))
	return addr
}

func (sp *Speaker) WriteUDP(d []byte) error {
	if sp.Conn == nil {
		// return fmt.Errorf("speaker %d not connected", sp.ID)
		return nil
	}
	n, err := sp.Conn.Write(d)
	if err != nil {
		sp.Statistic.Error += uint32(len(d))
		return fmt.Errorf("write to speaker '%d' failed: %s", sp.ID, err.Error())
	}

	sp.Statistic.Spend += uint64(n)

	if n != len(d) {
		sp.Statistic.Error += uint32(len(d) - n)
		return fmt.Errorf("write to speaker '%d' length error %d!=%d", sp.ID, n, len(d))
	}

	return nil
}

func (sp *Speaker) SetLine(newLine *Line) {
	if sp.Line != nil {
		sp.Line.RemoveSpeaker(sp)
	}

	ol := sp.Line
	sp.Line = newLine

	if newLine != nil {
		sp.LineId = newLine.ID

		newLine.AppendSpeaker(sp)
	} else {
		sp.LineId = 0
	}
	sp.SetChannel(audio.Channel_NONE)

	bus.DispatchObj(sp, "speaker edited", "line_id", sp.LineId)
	bus.DispatchObj(sp, "speaker line changed", ol)
}

func (l *Speaker) IsDeleted() bool {
	if l == nil {
		return true
	}
	return l.isDeleted
}

func (sp *Speaker) Save() {
	bus.DispatchObj(sp, "save speaker")
}

func (sp *Speaker) Elements() []stream.Element {
	return []stream.Element{
		sp.MixerEle,
		sp.EqualizerEle,
		sp.PlayerEle,
		sp.SpectrumEle,
		sp.VolumeEle,
	}
}

func (sp *Speaker) init() {
	sp.MixerEle = element.NewMixer()
	sp.VolumeEle = element.NewVolume(float64(sp.Volume) / 100)
	sp.SpectrumEle = element.NewSpectrum()
	sp.EqualizerEle = element.NewEqualizer(nil)
	sp.PlayerEle = element.NewPlayer()
	sp.PipeLine = pipeline.NewPipeLine(sp.Format(), sp.Elements()...)
}

func (o *Speaker) Dispatch(e string, args ...any) error {
	return bus.DispatchObj(o, e, args...)
}

func (o *Speaker) Register(e string, c func(o any, a ...any) error) *bus.HandlerData {
	return bus.RegisterObj(o, e, c)
}

func CountSpeaker() int {
	return len(speakerList)
}

func AllSpeakers() []*Speaker {
	return speakerList
}

func All(cb func(*Speaker)) {
	for _, sp := range speakerList {
		cb(sp)
	}
}

func NewSpeaker(ip string, line LineID, channel audio.Channel) (*Speaker, error) {
	locker.Lock()
	defer locker.Unlock()

	if s := FindSpeakerByIP(ip); s != nil {
		return s, nil
	}

	var sp Speaker

	sp.ID = getSpeakerID()
	sp.LineId = line
	sp.Ip = ip
	sp.Channel = uint32(channel)
	sp.State = State_OFFLINE
	sp.Volume = 50
	sp.init()

	speakerList = append(speakerList, &sp)

	if l := FindLineByID(line); l != nil {
		sp.Line = l
		l.AppendSpeaker(&sp)
	}

	sp.Save()

	BusSpeakerCreated.Dispatch(&sp)

	return &sp, nil
}

func DelSpeaker(id SpeakerID) error {
	sp := FindSpeakerByID(id)
	if sp == nil {
		return &UnknownSpeakerError{id}
	}

	locker.Lock()
	defer locker.Unlock()

	// 删除原始数据
	removeSpeaker(id)

	sp.Line.RemoveSpeaker(sp)
	sp.Line = nil

	sp.PipeLine.Close()

	BusSpeakerDeleted.Dispatch(sp)

	sp.PipeLine.Close()
	bus.UnregisterObj(sp)

	return nil
}

func removeSpeaker(id SpeakerID) {
	l := len(speakerList) - 1
	for i := 0; i <= l; i++ {
		if speakerList[i].ID == id {
			if i != l {
				speakerList[i] = speakerList[l]
			}
			speakerList = speakerList[:l]
		}
	}
}

func FindSpeakerByID(id SpeakerID) *Speaker {
	for _, sp := range speakerList {
		if sp.ID == id {
			return sp
		}
	}

	return nil
}

func getSpeakerID() (m SpeakerID) {
	m = maxSpeakerID

	for FindSpeakerByID(m) != nil {
		m++
		if m > SpeakerID_MAX {
			m = 1
		}
	}
	if m > maxSpeakerID && m < SpeakerID_MAX {
		maxSpeakerID = m + 1
	}

	return m
}

func FindSpeakerByIP(ip string) *Speaker {
	for i := 0; i < len(speakerList); i++ {
		if speakerList[i].Ip == ip {
			return speakerList[i]
		}
	}

	return nil
}
