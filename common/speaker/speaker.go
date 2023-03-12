package speaker

import (
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/element"
)

var speakerList []*Speaker

var lock sync.Mutex

type Speaker struct {
	Id        ID
	MAC       net.HardwareAddr
	IP        netip.Addr
	Name      string
	Line      *Line
	Mode      Model
	Dport     uint16 // pcm data port
	Supported bool   // 是否兼容

	RateMask    audio.AudioRateMask // 设备支持的采样率列表
	BitsMask    audio.BitsMask      // 设备支持的位宽列表
	Channel     audio.Channel       // 当前设置的声道
	AbsoluteVol bool                // 支持绝对音量控制
	PowerSave   bool                // 是否支持电源控制
	PowerSate   PowerState          // 电源状态

	Rate audio.Rate // 当前指定的采样率
	Bits audio.Bits // 当前指定的位宽

	Mixer     stream.MixerElement
	Player    stream.RawPlayerElement
	Volume    stream.VolumeElement // 音量
	Spectrum  stream.SpectrumElement
	Equalizer stream.EqualizerElement

	Conn  *net.UDPConn
	Queue chan QueueData

	Timeout    int // 超时计数
	ConnTime   time.Time
	State      State
	Statistic  Statistic
	LevelMeter float32
}

func (sp *Speaker) String() string {
	return sp.MAC.String()
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

func (sp *Speaker) SetOffline() {
	sp.State &= ^State_ONLINE
}

func (sp *Speaker) SetOnline() {
	sp.State |= State_ONLINE
}

func (sp *Speaker) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   sp.IP.AsSlice(),
		Zone: sp.IP.Zone(),
		Port: int(sp.Dport),
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
	if ch.IsValid() {
		sp.Channel = ch
	} else {
		sp.Channel = audio.Channel_NONE
	}
	sp.Line.refresh()
}

func (sp *Speaker) ChangeLine(newLine *Line) {
	sp.Line.RemoveSpeaker(sp)

	sp.Line = newLine

	newLine.AppendSpeaker(sp)
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

func NewSpeaker(id ID, line LineID, channel audio.Channel) (*Speaker, error) {
	lock.Lock()
	defer lock.Unlock()

	if s := FindSpeakerByID(id); s != nil {
		return s, nil
	}

	var sp Speaker
	sp.Id = id
	sp.Channel = channel
	sp.State = State_OFFLINE

	sp.Player = element.NewPlayer()
	sp.Mixer = element.NewMixer(sp.Player)
	sp.Volume = element.NewVolume(1)
	sp.Spectrum = element.NewSpectrum()
	sp.Equalizer = element.NewEqualizer(nil)

	speakerList = append(speakerList, &sp)

	if l := FindLineByID(line); l != nil {
		sp.Mixer.Add(l.Input.PipeLine)
		sp.Line = l
		l.AppendSpeaker(&sp)
	}

	return &sp, nil
}

func DelSpeaker(id ID) error {
	sp := FindSpeakerByID(id)
	if sp == nil {
		return &UnknownSpeakerError{id}
	}

	lock.Lock()
	defer lock.Unlock()

	// 删除原始数据
	removeSpeaker(id)

	sp.Line.RemoveSpeaker(sp)
	sp.Line = nil

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
	for i := 0; i < len(speakerList); i++ {
		if speakerList[i].Id == id {
			return speakerList[i]
		}
	}

	return nil
}
