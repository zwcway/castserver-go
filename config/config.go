package config

import (
	"fmt"
	"math"
	"net"
	"net/netip"
	"path/filepath"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/fasthttp-upnp/utils"
)

type Interface struct {
	Iface    *net.Interface
	AddrPort netip.AddrPort
	IPV6     bool
}

func (f Interface) String() string {
	if f.Iface != nil {
		return f.Iface.Name
	}
	return f.AddrPort.String()
}

var (
	ConfigFile string = "castserver.conf"

	RuntimeThreads int

	ServerListen     Interface  // 监听的地址
	MulticastAddress netip.Addr // 多播的地址
	MulticastPort    uint16     // 多播端口
	ServerNetMTU     uint32

	ReadBufferSize  int
	ReadQueueSize   int
	SendRoutinesMax int
	SendQueueSize   int

	SupportAudioBits  []audio.Bits
	SupportAudioRates []audio.Rate

	AudioBuferSize int

	SpeakerOfflineTimeout       int
	SpeakerOfflineCheckInterval int

	HTTPListen Interface
	HTTPNetMTU uint32
	HTTPRoot   string

	ReceiveListen  Interface
	ReceiveTempDir string
	EnableDLNA     bool
	EnableAirPlay  bool

	DLNAListen         Interface
	DLNANotifyInterval uint8
	DLNAAllowIps       []*net.IPNet
	DLNADenyIps        []*net.IPNet
)

func MTU() int {
	if ServerListen.Iface == nil {
		return int(ServerNetMTU)
	}
	return ServerListen.Iface.MTU
}

func OfflineValue() int {
	return int(math.Ceil(float64(SpeakerOfflineTimeout) / float64(SpeakerOfflineCheckInterval)))
}

func NameVersion() string {
	return fmt.Sprintf("%s %s", APPNAME, VERSION)
}

func TemporayFile(file string) string {
	if ReceiveTempDir == "" {
		return ""
	}
	// if strings.HasPrefix(file, "http") {
	// 	idx := strings.Index(file, "?")
	// 	file = file[:idx]
	// }
	ext := filepath.Ext(file)

	return filepath.Join(ReceiveTempDir, utils.MakeUUID(file)+ext)
}

const (
	APPNAME string = "Castspeaker Server"
	VERSION string = "1.0"
)
