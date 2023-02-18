package speaker

import (
	"net"
	"net/netip"
	"sync"

	"github.com/zwcway/castserver-go/common/audio"
)

type Speaker struct {
	ID        SpeakerID
	MAC       net.HardwareAddr
	IP        netip.Addr
	Name      string
	Line      SpeakerLineID
	Mode      SpeakerModel
	Dport     uint16 // pcm data port
	Mport     uint16 // multicast port
	Supported bool

	RateMask audio.AudioRateMask
	BitsMask audio.AudioBitsMask
	Channel  audio.AudioChannel

	Conn *net.UDPConn

	Timeout    int
	ConnTime   uint64
	State      SpeakerState
	Statistic  SpeakerStatistic
	LevelMeter float32
}

func (sp *Speaker) String() string {
	return sp.MAC.String()
}

func (sp *Speaker) IsOnline() bool {
	return sp.State == SpeakerState_ONLINE
}
func (sp *Speaker) IsOffline() bool {
	return sp.State == SpeakerState_OFFLINE
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
	sp.State &= ^SpeakerState_ONLINE
}

func (sp *Speaker) SetOnline() {
	sp.State |= SpeakerState_ONLINE
}

func (sp *Speaker) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   sp.IP.AsSlice(),
		Zone: sp.IP.Zone(),
		Port: int(sp.Dport),
	}
}

type speakerMapSlice map[int][]*Speaker

func (s *speakerMapSlice) remove(key int, sp *Speaker) {
	slice, ok := (*s)[key]
	if !ok {
		return
	}
	for i, item := range slice {
		if item == sp {
			(*s)[key] = append(slice[:i], slice[i+1:]...)
			return
		}
	}
}
func (s *speakerMapSlice) add(key int, sp *Speaker) {
	if _, ok := (*s)[key]; ok {
		(*s)[key] = append((*s)[key], sp)
	} else {
		(*s)[key] = append([]*Speaker{}, sp)
	}
}

var list []Speaker
var listByID map[SpeakerID]*Speaker
var listByLine speakerMapSlice
var listByChannel speakerMapSlice

var lineList map[SpeakerLineID]SpeakerLine

var lock sync.Mutex

func Init() error {
	maxSize := 0

	list = make([]Speaker, maxSize)
	listByID = make(map[SpeakerID]*Speaker, 0)
	listByLine = make(speakerMapSlice, 0)
	listByChannel = make(speakerMapSlice, 0)

	lineList = make(map[SpeakerLineID]SpeakerLine, 0)

	return nil
}

func lineIsValid(line SpeakerLineID) bool {
	_, ok := lineList[line]
	return ok
}

func AddLine(name string) *SpeakerLine {
	var line SpeakerLine

	l := len(lineList)

	if l > 0 {
		line.id = lineList[SpeakerLineID(l-1)].id + 1
	} else {
		line.id = SpeakerLineID(0)
	}
	line.name = name

	lineList[line.id] = line

	return &line
}

func EditLine(id SpeakerLineID, name string) error {
	line, ok := lineList[id]
	if !ok {
		return &UnknownLineError{id}
	}
	line.name = name

	return nil
}

func DelLine(id SpeakerLineID, move SpeakerLineID) error {
	lock.Lock()
	defer lock.Unlock()

	if id == move {
		return nil
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

func CountLineSpeaker(id SpeakerLineID) int {
	i := 0
	for _, sp := range listByID {
		if sp.Line == id {
			i++
		}
	}
	return i
}

func All(cb func(*Speaker)) {
	for _, sp := range list {
		cb(&sp)
	}
}

func AddSpeaker(id SpeakerID, line SpeakerLineID, channel audio.AudioChannel) (*Speaker, error) {
	lock.Lock()
	defer lock.Unlock()

	if s := FindSpeakerByID(id); s != nil {
		return s, nil
	}

	var sp Speaker
	sp.ID = id
	sp.Line = line
	sp.Channel = channel
	sp.State = SpeakerState_OFFLINE

	list = append(list, sp)

	listByID[sp.ID] = &sp
	listByLine.add(int(sp.Line), &sp)
	listByChannel.add(int(sp.Channel), &sp)

	return &sp, nil
}

func DelSpeaker(id SpeakerID) error {
	sp, ok := listByID[id]
	if !ok {
		return &UnknownSpeakerError{id}
	}

	lock.Lock()
	defer lock.Unlock()

	// 删除原始数据
	for i, s := range list {
		if &s == sp {
			list[i] = list[len(list)-1]
			list = list[:len(list)-1]
			break
		}
	}

	delete(listByID, sp.ID)
	listByLine.remove(int(sp.Line), sp)
	listByChannel.remove(int(sp.Channel), sp)

	return nil
}

func FindSpeakerByID(id SpeakerID) *Speaker {
	if sp, ok := listByID[id]; ok {
		return sp
	}
	return nil
}

func SpeakersByChannel(ch audio.AudioChannel) ([]*Speaker, bool) {
	l, ok := listByChannel[int(ch)]
	return l, ok
}
func SpeakersByLine(line SpeakerLineID) ([]*Speaker, bool) {
	l, ok := listByLine[int(line)]
	return l, ok
}
