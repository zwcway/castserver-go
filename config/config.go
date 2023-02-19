package config

import (
	"fmt"
	"math"
	"net"
	"net/netip"
	"path/filepath"

	"github.com/zwcway/fasthttp-upnp/utils"
)

var (
	ConfigFile string = "castserver.conf"

	DetectUseIPV6    bool
	DetectInterface  *net.Interface
	MulticastAddress netip.Addr
	MulticastPort    uint16
	DetectNetMTU     uint32

	MaxReadBufferSize int
	SendRoutinesMax   int
	SendQueueSize     int

	SupportAudioBits  []uint8
	SupportAudioRates []uint8

	SpeakerOfflineTimeout       int
	SpeakerOfflineCheckInterval int

	HTTPUseIPV6   bool
	HTTPInterface *net.Interface
	HTTPAddrPort  string
	HTTPNetMTU    uint32
	HTTPRoot      string

	ReceiveUseIPV6   bool
	ReceiveAddrPort  netip.AddrPort
	ReceiveInterface *net.Interface
	ReceiveTempDir   string
	EnableDLNA       bool
	EnableAirPlay    bool

	DLNAUseIPV6        bool
	DLNAInterface      *net.Interface
	DLNAAddrPort       netip.AddrPort
	DLNANotifyInterval uint8
	DLNAAllowIps       []*net.IPNet
	DLNADenyIps        []*net.IPNet
)

func MTU() int {
	if DetectInterface == nil {
		return int(DetectNetMTU)
	}
	return DetectInterface.MTU
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
