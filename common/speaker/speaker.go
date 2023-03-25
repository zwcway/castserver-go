package speaker

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/stream"
)

var speakerList []*Speaker

type SpeakerConfig struct {
	MAC   net.HardwareAddr
	IP    netip.Addr
	ipStr string
	Mode  Model
	Dport uint16 // pcm data port

	RateMask    audio.AudioRateMask // 设备支持的采样率列表
	BitsMask    audio.BitsMask      // 设备支持的位宽列表
	AbsoluteVol bool                // 支持绝对音量控制
	PowerSave   bool                // 是否支持电源控制
}

func (sc *SpeakerConfig) SetIP(ip netip.Addr) {
	sc.IP = ip
	sc.ipStr = ip.String()
}

type Speaker struct {
	Id        ID
	Name      string
	Line      *Line
	Supported bool // 是否兼容

	Channel audio.Channel // 当前设置的声道
	Rate    audio.Rate    // 当前指定的采样率
	Bits    audio.Bits    // 当前指定的位宽

	PowerSate PowerState // 电源状态

	Config SpeakerConfig

	PipeLine  stream.PipeLiner
	Mixer     stream.MixerElement
	Player    stream.RawPlayerElement
	Volume    stream.VolumeElement // 音量
	Spectrum  stream.SpectrumElement
	Equalizer stream.EqualizerElement
	Resample  stream.ResampleElement

	Conn  *net.UDPConn
	Queue chan QueueData

	Timeout   int // 超时计数
	ConnTime  time.Time
	State     State
	Statistic Statistic

	isDeleted bool
}

func (sp *Speaker) String() string {
	return fmt.Sprintf("%s(%s)", sp.Config.IP.String(), sp.Config.MAC.String())
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
	if sp.Config.Dport > 0 {
		sp.SetOnline()
	} else {
		sp.SetOffline()
	}
}

func (sp *Speaker) SetOffline() {
	sp.State &= ^State_ONLINE
}

func (sp *Speaker) SetOnline() {
	sp.State |= State_ONLINE
}

func (sp *Speaker) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   sp.Config.IP.AsSlice(),
		Zone: sp.Config.IP.Zone(),
		Port: int(sp.Config.Dport),
	}
}

func (sp *Speaker) WriteUDP(d []byte) error {
	if sp.Conn == nil {
		return fmt.Errorf("speaker %d not connected", sp.Id)
	}
	n, err := sp.Conn.Write(d)
	if err != nil {
		sp.Statistic.Error += uint32(len(d))
		return fmt.Errorf("write to speaker '%d' failed: %s", sp.Id, err.Error())
	}

	sp.Statistic.Spend += uint64(n)

	if n != len(d) {
		sp.Statistic.Error += uint32(len(d) - n)
		return fmt.Errorf("write to speaker '%d' length error %d!=%d", sp.Id, n, len(d))
	}

	return nil
}

func (sp *Speaker) ChangeChannel(ch audio.Channel) {
	och := sp.Channel

	if ch.IsValid() {
		sp.Channel = ch
	} else {
		sp.Channel = audio.Channel_NONE
	}
	sp.Line.refresh()

	bus.Trigger("speaker channel moved", sp, och)
}

func (sp *Speaker) ChangeLine(newLine *Line) {
	sp.Line.RemoveSpeaker(sp)

	ol := sp.Line
	sp.Line = newLine

	newLine.AppendSpeaker(sp)

	bus.Trigger("speaker line changed", sp, ol)
}

func (sp *Speaker) Format() audio.Format {
	return audio.Format{
		SampleRate: sp.Rate,
		Layout:     audio.NewChannelLayout(sp.Channel),
		SampleBits: sp.Bits,
	}
}

func (l *Speaker) IsDeleted() bool {
	return l.isDeleted
}

func initSpeaker() error {
	maxSize := 0

	speakerList = make([]*Speaker, maxSize)

	return nil
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

	sp.Id = getSpeakerID()
	sp.Channel = channel
	sp.State = State_OFFLINE

	sp.Player = element.NewPlayer()
	sp.Mixer = element.NewMixer(sp.Player)
	sp.Volume = element.NewVolume(1)
	sp.Spectrum = element.NewSpectrum()
	sp.Equalizer = element.NewEqualizer(nil)
	sp.Resample = element.NewResample(sp.Format())
	sp.PipeLine = pipeline.NewPipeLine(sp.Format(),
		sp.Mixer,
		sp.Equalizer,
		sp.Spectrum,
		sp.Volume,
		sp.Resample,
	)

	speakerList = append(speakerList, &sp)

	if l := FindLineByID(line); l != nil {
		sp.Line = l
		l.AppendSpeaker(&sp)
	}

	bus.Trigger("speaker created", &sp)

	return &sp, nil
}

func DelSpeaker(id ID) error {
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

	bus.Trigger("speaker deleted", sp)

	return nil
}

func removeSpeaker(id ID) {
	l := len(speakerList) - 1
	for i := 0; i <= l; i++ {
		if speakerList[i].Id == id {
			if i != l {
				speakerList[i] = speakerList[l]
			}
			speakerList = speakerList[:l]
		}
	}
}

func FindSpeakerByID(id ID) *Speaker {
	for _, sp := range speakerList {
		if sp.Id == id {
			return sp
		}
	}

	return nil
}

func getSpeakerID() (m ID) {
	m = maxSpeakerID

	for FindSpeakerByID(m) != nil {
		m++
		if m > ID_MAX {
			m = 1
		}
	}
	if m > maxSpeakerID && m < ID_MAX {
		maxSpeakerID = m + 1
	}

	return m
}

func FindSpeakerByIP(ip string) *Speaker {
	for i := 0; i < len(speakerList); i++ {
		if speakerList[i].Config.ipStr == ip {
			return speakerList[i]
		}
	}

	return nil
}
