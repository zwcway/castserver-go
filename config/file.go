package config

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/zwcway/castserver-go/utils"

	"github.com/zwcway/castserver-go/common/audio"

	"github.com/go-ini/ini"
	"go.uber.org/zap"
)

func parseListen(log *zap.Logger, listen string, port uint16) (ap netip.AddrPort, iface *net.Interface) {
	if len(listen) != 0 {
		var err error
		// 支持 IP:PORT 或 [IPV6]:PORT 格式
		ap, err = netip.ParseAddrPort(listen)
		if err == nil {
			return
		}

		// 支持仅 IP 或  IPV6 格式
		addr, err := netip.ParseAddr(listen)
		if err == nil {
			ap = netip.AddrPortFrom(addr, port)
			return
		}

		iface = utils.InterfaceByName(listen)
		if iface == nil {
			log.Error("this is not a interface name", zap.String("listen", listen))
		}
	}
	ap = netip.AddrPortFrom(netip.MustParseAddr("0.0.0.0"), port)

	return
}

func parseDetectListen(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	port, err := def.Uint()
	if err != nil {
		panic(err)
	}
	var listen string
	if k != nil {
		listen = k.String()
	}
	addr, iface := parseListen(log, listen, uint16(port))

	ServerUseIPV6 = addr.Addr().Is6()
	ServerAddrPort = addr
	ServerInterface = iface
}
func parseHttpListen(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	port, err := def.Uint()
	if err != nil {
		panic(err)
	}
	var listen string
	if k != nil {
		listen = k.String()
	}
	addr, iface := parseListen(log, listen, uint16(port))

	HTTPUseIPV6 = addr.Addr().Is6()
	HTTPAddrPort = addr.String()
	HTTPInterface = iface
}
func parseReceiveListen(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	port, err := def.Uint()
	if err != nil {
		port = 0
	}
	var listen string
	if k != nil {
		listen = k.String()
	}

	addr, iface := parseListen(log, listen, uint16(port))

	ReceiveUseIPV6 = addr.IsValid() && addr.Addr().Is6()
	ReceiveAddrPort = addr
	ReceiveInterface = iface
}

func parseDLNAListen(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	port, err := def.Uint()
	if err != nil {
		port = 0
	}
	var listen string
	if k != nil {
		listen = k.String()
	}

	addr, iface := parseListen(log, listen, uint16(port))

	DLNAUseIPV6 = addr.IsValid() && addr.Addr().Is6()
	DLNAAddrPort = addr
	DLNAInterface = iface
}

func parseBits(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	parse := func(b string) bool {
		r := strings.FieldsFunc(b, func(r rune) bool {
			return r == ' ' || r == '|' || r == '/'
		})

		SupportAudioBits = make([]uint8, 0)
		for _, v := range r {
			var a audio.Bits
			a = a.FromName(v)
			if a != audio.AudioBits_NONE {
				SupportAudioBits = append(SupportAudioBits, uint8(a))
				continue
			}
			i, err := strconv.ParseInt(v, 0, 1)
			if err != nil {
				return false
			}
			a = audio.NewAudioBits(int32(i))
			if !a.IsValid() {
				return false
			}
			SupportAudioBits = append(SupportAudioBits, uint8(a))
		}
		return true
	}
	if k != nil {
		b := k.String()
		if len(b) > 0 && parse(b) {
			return
		}
	}
	parse(def.String())
}

func parseRates(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	parse := func(b string) bool {
		r := strings.FieldsFunc(b, func(r rune) bool {
			return r == ' ' || r == '|' || r == '/'
		})

		SupportAudioRates = make([]uint8, 0)
		for _, v := range r {
			i, err := strconv.ParseInt(v, 0, 32)
			if err != nil {
				return false
			}
			a := audio.NewAudioRate(int(i))
			if !a.IsValid() {
				return false
			}
			SupportAudioRates = append(SupportAudioRates, uint8(a))
		}
		return true
	}
	if k != nil {
		b := k.String()

		if len(b) > 0 && parse(b) {
			return
		}
	}
	parse(def.String())
}

func parsePath(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	path := def.String()
	if k != nil {
		path = k.String()
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error("path not exists", zap.String("path", path), zap.String("key", ck.key))
		cfg.SetString(def.String())
	} else {
		cfg.SetString(path)
	}
}
func parseTempDir(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey) {
	path := def.String()
	if k != nil {
		path = k.String()
	}

	if path == "" {
		path = filepath.Join(os.TempDir(), APPNAME)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModeTemporary)
		if err != nil {
			log.Error("can not create temp dir ", zap.String("dir", path))
			path = ""
		}
	}

	cfg.SetString(path)
}

type cfgKey struct {
	cfg any
	key string
	def string
	cb  func(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *cfgKey)
}
type cfgSection struct {
	name string
	keys []cfgKey
}

func FromFile(log *zap.Logger, file string) error {
	data, err := os.ReadFile(file)

	if err != nil {
		return err
	}

	ConfigFile = file

	return FromContent(log, data)
}

func FromContent(log *zap.Logger, data []byte) error {
	c, err := ini.Load(data)
	if err != nil {
		return err
	}

	for _, cs := range configStruct {
		section := c.Section(cs.name)

		for _, ck := range cs.keys {
			key, _ := section.GetKey(ck.key)
			defKey, _ := section.NewKey(ck.key+"-default", ck.def)

			isNil := ck.cfg == nil

			cfgrv := reflect.ValueOf(ck.cfg)

			if !isNil && cfgrv.Kind() != reflect.Pointer {
				panic(fmt.Errorf("the '%s' must be a pointer", reflect.TypeOf(ck.cfg).Name()))
			}

			if !isNil {
				cfgrv = cfgrv.Elem()
			}

			if ck.cb != nil {
				ck.cb(log, cfgrv, key, defKey, &ck)
				continue
			}

			if isNil {
				continue
			}

			switch cfgrv.Kind() {
			case reflect.String:
				setReflectString(log, key, defKey, cfgrv, &ck)
			case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8, reflect.Int64:
				setReflectInt(log, key, defKey, cfgrv, &ck)
			case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
				setReflectUInt(log, key, defKey, cfgrv, &ck)
			case reflect.Bool:
				setReflectBool(log, key, defKey, cfgrv, &ck)
			case reflect.Struct:
				setReflectStruct(log, key, defKey, cfgrv, &ck)
			case reflect.Slice:
				setReflectSlice(log, key, defKey, cfgrv, &ck)
			}
		}
	}

	if ServerUseIPV6 {
		MulticastAddress = netip.MustParseAddr("FF02:2C:4D:FF::16")
	} else {
		MulticastAddress = netip.MustParseAddr("239.44.77.16")
	}
	MulticastPort = 4414

	return nil
}

func setReflectString(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *cfgKey) {
	if key == nil {
		cfgrv.SetString(ck.def)
	} else {
		i := key.String()
		if key == nil || len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
			i = ck.def
		}
		cfgrv.SetString(i)
	}
}

func setReflectInt(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *cfgKey) {
	i, err := defKey.Int()
	if err != nil {
		panic(nil)
	}

	if key != nil {
		if len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
		} else if ki, err := key.Int(); err != nil {
			log.Error("value invalid", zap.String("key", ck.key), zap.String("val", key.String()))
			i = ki
		}
	}
	cfgrv.SetInt(int64(i))

}
func setReflectUInt(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *cfgKey) {
	i, err := defKey.Uint()
	if err != nil {
		panic(nil)
	}
	if key != nil {
		if len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
		} else if ki, err := key.Uint(); err != nil {
			log.Error("value invalid", zap.String("key", ck.key), zap.String("val", key.String()))
			i = ki
		}
	}

	cfgrv.SetUint(uint64(i))

}
func setReflectBool(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *cfgKey) {
	i, err := defKey.Bool()
	if err != nil {
		panic(nil)
	}
	if key != nil {
		if len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
		} else if ki, err := key.Bool(); err != nil {
			log.Error("value invalid", zap.String("key", ck.key), zap.String("val", key.String()))
			i = ki
		}
	}
	cfgrv.SetBool(i)
}
func setReflectStruct(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *cfgKey) {
	var keyV string
	isEmpty := key != nil && len(key.String()) == 0

	if key == nil {
		keyV = defKey.String()
	} else {
		keyV = key.String()
	}
	if isEmpty {
		log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
	}
	as := cfgrv.Type().AssignableTo

	if as(reflect.PointerTo(reflect.TypeOf(net.Interface{}))) {
		if len(keyV) == 0 {
			cfgrv.Set(reflect.Zero(cfgrv.Type()))
			return
		}
		ifi := utils.InterfaceByName(keyV)
		if ifi == nil {
			log.Error("this is not a interface name", zap.String("name", keyV))
		}
		cfgrv.Set(reflect.ValueOf(ifi))
	}
}

func setReflectSlice(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *cfgKey) {
	var keyV string
	isEmpty := key != nil && len(key.String()) == 0

	if key == nil {
		keyV = defKey.String()
	} else {
		keyV = key.String()
	}
	if isEmpty {
		log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
	}
	elemRT := cfgrv.Type().Elem()
	as := elemRT.AssignableTo

	if as(reflect.PointerTo(reflect.TypeOf(net.IPNet{}))) {
		if len(keyV) == 0 {
			cfgrv.Set(reflect.Zero(cfgrv.Type()))
			return
		}
		keyList := strings.FieldsFunc(keyV, func(r rune) bool {
			return r == ' ' || r == '|'
		})
		for _, ipstr := range keyList {
			_, ipnet, err := net.ParseCIDR(ipstr)
			if err != nil {
				ip := net.ParseIP(ipstr)
				if ip == nil {
					log.Error("ip invalid", zap.String("key", ck.key), zap.String("val", key.String()))
					return
				}
				var m net.IPMask
				if strings.ContainsRune(ipstr, '.') {
					m = net.CIDRMask(32, 8*net.IPv4len)
				} else {
					m = net.CIDRMask(128, 8*net.IPv6len)
				}
				ipnet = &net.IPNet{IP: ip.Mask(m), Mask: m}
			}
			ele := reflect.New(elemRT.Elem())
			ele.Elem().Set(reflect.ValueOf(ipnet).Elem())
			cfgrv.Set(reflect.Append(cfgrv, ele))
		}
	}
}

// TODO 增加配置项的通用校验规则

var configStruct = []cfgSection{
	{"", []cfgKey{
		{&ServerNetMTU, "mtu", "1500", nil},
		{&RuntimeThreads, "max thread", "100", nil},
	}},
	{"audio", []cfgKey{
		{&SupportAudioBits, "support bits", "u8/u16le/u24le/s32le/u32le/fltle", parseBits},
		{&SupportAudioRates, "support rates", "44100/48000/96000/192000", parseRates},
	}},
	{"detect", []cfgKey{
		{nil, "listen", "4414", parseDetectListen},
		{&SpeakerOfflineTimeout, "offline timeout", "5", nil},
		{&SpeakerOfflineCheckInterval, "offline check interval", "5", nil},
	}},
	{"speaker", []cfgKey{
		{&ReadBufferSize, "receive buffer", "1024", nil},
		{&SendRoutinesMax, "send thread max", "2", nil},
		{&SendQueueSize, "send queue size", "16", nil},
		{&ReadQueueSize, "read queue size", "512", nil},
	}},
	{"http", []cfgKey{
		{nil, "listen", "4415", parseHttpListen},
		{&HTTPRoot, "root", "web/public", parsePath},
	}},
	{"receive", []cfgKey{
		{nil, "listen", "4416", parseReceiveListen},
		{&ReceiveTempDir, "tempdir", "", parseTempDir},
		{&EnableDLNA, "dlna", "true", nil},
	}},
	{"dlna", []cfgKey{
		{nil, "listen", "4416", parseDLNAListen},
		{&DLNANotifyInterval, "notify interval", "30", nil},
		{&DLNAAllowIps, "allow ips", "", nil},
		{&DLNADenyIps, "deny ips", "", nil},
	}},
}
