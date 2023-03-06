package speaker

import (
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/dsp"
)

type Speaker struct {
	ID        ID
	MAC       net.HardwareAddr
	IP        netip.Addr
	Name      string
	Line      LineID
	Mode      Model
	Dport     uint16 // pcm data port
	Supported bool   // 是否兼容

	RateMask audio.AudioRateMask // 设备支持的采样率列表
	BitsMask audio.BitsMask      // 设备支持的位宽列表
	Channel  audio.Channel       // 当前设置的声道
	Rate     audio.Rate          // 当前指定的采样率
	Bits     audio.Bits          // 当前指定的位宽

	AbsoluteVol bool // 支持绝对音量控制
	Volume      int  // 音量
	IsMute      bool

	PowerSave bool       // 是否支持电源控制
	PowerSate PowerState // 电源状态

	Conn *net.UDPConn

	Timeout    int // 超时计数
	ConnTime   time.Time
	State      State
	Statistic  Statistic
	LevelMeter float32

	DPEnable        bool
	dsp.DataProcess // 数字频域均衡器
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
		return fmt.Errorf("speaker %d not connected", sp.ID)
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

func (sp *Speaker) ChangeChannel(ch audio.Channel) {
	if ch.IsValid() {
		sp.Channel = ch
	} else {
		sp.Channel = audio.AudioChannel_NONE
	}
	refreshLine(sp.Line)
}

func (sp *Speaker) ChangeLine(line LineID) {
	removeSpeakerFromLine(sp)

	sp.Line = line

	appendSpeakerToLine(sp)
	refreshLine(line)
}

var list []*Speaker
var listByID map[ID]*Speaker

var lock sync.Mutex

func Init() error {
	maxSize := 0

	list = make([]*Speaker, maxSize)
	listByID = make(map[ID]*Speaker, 0)

	initLine()

	return nil
}

func CountSpeaker() int {
	return len(list)
}

func AllSpeakers() []*Speaker {
	return list
}

func All(cb func(*Speaker)) {
	for _, sp := range list {
		cb(sp)
	}
}

func AddSpeaker(id ID, line LineID, channel audio.Channel) (*Speaker, error) {
	lock.Lock()
	defer lock.Unlock()

	if s := FindSpeakerByID(id); s != nil {
		return s, nil
	}

	var sp Speaker
	sp.ID = id
	sp.Line = line
	sp.Channel = channel
	sp.State = State_OFFLINE

	list = append(list, &sp)

	listByID[sp.ID] = &sp
	appendSpeakerToLine(&sp)

	return &sp, nil
}

func DelSpeaker(id ID) error {
	sp, ok := listByID[id]
	if !ok {
		return &UnknownSpeakerError{id}
	}

	lock.Lock()
	defer lock.Unlock()

	// 删除原始数据
	for i, s := range list {
		if s == sp {
			list[i] = list[len(list)-1]
			list = list[:len(list)-1]
			break
		}
	}

	delete(listByID, sp.ID)
	removeSpeakerFromLine(sp)

	return nil
}

func FindSpeakerByID(id ID) *Speaker {
	if sp, ok := listByID[id]; ok {
		return sp
	}
	return nil
}
