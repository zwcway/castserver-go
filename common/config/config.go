package config

import (
	"fmt"
	"math"
	"net"
	"net/netip"
	"path/filepath"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
	upnputils "github.com/zwcway/fasthttp-upnp/utils"
	"gorm.io/gorm"
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
	log lg.Logger

	ConfigFile string = "castserver.conf"

	RuntimeThreads int = 100

	// 多播的地址
	MulticastAddress netip.Addr = netip.MustParseAddr("239.44.77.16")
	MulticastPort    uint16     = 4414 // 多播端口

	ServerNetMTU uint32 = 1500
	// 监听的地址
	ServerListen Interface = Interface{
		AddrPort: netip.AddrPortFrom(netip.MustParseAddr("0.0.0.0"), MulticastPort),
	}

	DB *gorm.DB

	ReadBufferSize  int = 1024
	ReadQueueSize   int = 512
	SendRoutinesMax int = 2
	SendQueueSize   int = 16

	SupportAudioBits []audio.Bits = []audio.Bits{
		audio.Bits_U8,
		audio.Bits_U16LE,
		audio.Bits_U24LE,
		audio.Bits_U32LE,
		audio.Bits_S32LE,
		audio.Bits_32LEF,
	}
	SupportAudioRates []audio.Rate = []audio.Rate{
		audio.AudioRate_44100,
		audio.AudioRate_48000,
		audio.AudioRate_96000,
		audio.AudioRate_192000,
	}

	// 缓冲长度（单位ms），0 表示动态自动判断
	AudioBuferMSDuration MilliDuration = 10 * time.Millisecond

	SpeakerOfflineTimeout       int = 5
	SpeakerOfflineCheckInterval int = 5

	// tcp
	HTTPListen Interface = Interface{
		AddrPort: netip.MustParseAddrPort("0.0.0.0:4415"),
	}
	HTTPRoot    string = "web/public"
	WSClientMAX int    = 5

	// udp
	ReceiveListen Interface = Interface{
		AddrPort: netip.MustParseAddrPort("0.0.0.0:4416"),
	}
	ReceiveTempDir string = ""
	EnableDLNA     bool   = false
	EnableAirPlay  bool   = false

	// tcp
	DLNAListen Interface = Interface{
		AddrPort: netip.MustParseAddrPort("0.0.0.0:4416"),
	}
	DLNANotifyInterval uint8        = 30
	DLNAAllowIps       []*net.IPNet = []*net.IPNet{}
	DLNADenyIps        []*net.IPNet = []*net.IPNet{}
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

	return filepath.Join(ReceiveTempDir, upnputils.MakeUUID(file)+ext)
}

func initLogger(ctx utils.Context) {
	if log == nil {
		log = ctx.Logger("config")
	}
}

const (
	APPNAME string = "Castspeaker Server"
	VERSION string = "1.0"
)
