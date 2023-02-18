package config

import (
	"github.com/zwcway/castserver-go/utils"
	"fmt"
	"net"
	"net/netip"
	"os"
	"reflect"
	"strconv"
	"strings"

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

func parseHttpListen(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key) {
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
func parseReceiveListen(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key) {
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

func parseBits(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key) {
	parse := func(b string) bool {
		r := strings.FieldsFunc(b, func(r rune) bool {
			return r == ' ' || r == '|' || r == '/'
		})

		SupportAudioBits = make([]uint8, 0)
		for _, v := range r {
			i, err := strconv.ParseInt(v, 0, 1)
			if err != nil {
				return false
			}
			a, err := audio.NewAudioBits(int32(i))
			if err != nil {
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

func parseRates(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key) {
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
			a, err := audio.NewAudioRate(int32(i))
			if err != nil {
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

func parsePath(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key) {
	path := def.String()
	if k != nil {
		path = k.String()
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error("path not exists", zap.String("path", path), zap.String("key", k.Name()))
		cfg.SetString(def.String())
	} else {
		cfg.SetString(path)
	}
}

type cfgKey struct {
	cfg any
	key string
	def string
	cb  func(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key)
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

			isEmpty := key == nil || len(key.String()) == 0
			isNil := ck.cfg == nil

			cfgrv := reflect.ValueOf(ck.cfg)

			if !isNil && cfgrv.Kind() != reflect.Pointer {
				panic(fmt.Errorf("the '%s' must be a pointer", reflect.TypeOf(ck.cfg).Name()))
			}

			if !isNil {
				cfgrv = cfgrv.Elem()
			}

			if ck.cb != nil {
				ck.cb(log, cfgrv, key, defKey)
				continue
			}

			if isNil {
				continue
			}

			switch cfgrv.Kind() {
			case reflect.String:
				if key == nil {
					cfgrv.SetString(ck.def)
				} else {
					i := key.String()
					if isEmpty {
						log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
						i = ck.def
					}
					cfgrv.SetString(i)
				}
			case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8, reflect.Int64:
				i, err := defKey.Int()
				if err != nil {
					panic(nil)
				}
				if key != nil {
					if isEmpty {
						log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
					} else if ki, err := key.Int(); err != nil {
						log.Error("value invalid", zap.String("key", ck.key), zap.String("val", key.String()))
						i = ki
					}
				}
				cfgrv.SetInt(int64(i))
			case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
				i, err := defKey.Uint()
				if err != nil {
					panic(nil)
				}
				if key != nil {
					if isEmpty {
						log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
					} else if ki, err := key.Uint(); err != nil {
						log.Error("value invalid", zap.String("key", ck.key), zap.String("val", key.String()))
						i = ki
					}
				}
				cfgrv.SetUint(uint64(i))
			case reflect.Bool:
				i, err := defKey.Bool()
				if err != nil {
					panic(nil)
				}
				if key != nil {
					if isEmpty {
						log.Error("value empty", zap.String("key", ck.key), zap.String("val", key.String()))
					} else if ki, err := key.Bool(); err != nil {
						log.Error("value invalid", zap.String("key", ck.key), zap.String("val", key.String()))
						i = ki
					}
				}
				cfgrv.SetBool(i)
			case reflect.Struct:
				var keyV string
				if key == nil || isEmpty {
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
						break
					}
					ifi := utils.InterfaceByName(keyV)
					if ifi == nil {
						log.Error("this is not a interface name", zap.String("name", keyV))
					}
					cfgrv.Set(reflect.ValueOf(ifi))
				}
			}
		}
	}

	if DetectUseIPV6 {
		MulticastAddress = netip.MustParseAddr("FF02:2C:4D:FF::16")
	} else {
		MulticastAddress = netip.MustParseAddr("239.44.77.16")
	}
	MulticastPort = 4414

	return nil
}

// TODO 增加配置项的通用校验规则

var configStruct = []cfgSection{
	{"", []cfgKey{
		{&DetectNetMTU, "mtu", "1500", nil},
	}},
	{"audio", []cfgKey{
		{&SupportAudioBits, "support bits", "u8/u16le/u24le/s32le/u32le/fltle", parseBits},
		{&SupportAudioRates, "support rates", "44100/48000/96000/192000", parseRates},
	}},
	{"detect", []cfgKey{
		{&DetectUseIPV6, "ipv6", "false", nil},
		{&DetectInterface, "interface", "", nil},
		{&MaxReadBufferSize, "read buffer", "1024", nil},
		{&SpeakerOfflineTimeout, "offline timeout", "1024", nil},
		{&SpeakerOfflineCheckInterval, "offline check interval", "1024", nil},
	}},
	{"push", []cfgKey{
		{&SendRoutinesMax, "send routines max", "2", nil},
		{&SendQueueSize, "send queue size", "512", nil},
	}},
	{"http", []cfgKey{
		{nil, "listen", "4415", parseHttpListen},
		{&HTTPRoot, "root", "front/public", parsePath},
	}},
	{"receive", []cfgKey{
		{nil, "listen", "4416", parseReceiveListen},
		{&EnableDLNA, "dlna", "true", nil},
	}},
	{"dlna", []cfgKey{
		{&DLNANotifyInterval, "notify interval", "30", nil},
	}},
}
