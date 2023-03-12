package config

import (
	"fmt"
	"io"
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

const (
	SEP     = '|'
	Unknown = "unknown"
)

func parseListen(log *zap.Logger, cfg reflect.Value, listen string, port uint16) {
	var li Interface
	if len(listen) != 0 {
		var err error
		// 支持 IP:PORT 或 [IPV6]:PORT 格式
		li.AddrPort, err = netip.ParseAddrPort(listen)
		if err == nil {
			goto _default_
		}

		// 支持仅 IP 或  IPV6 格式
		addr, err := netip.ParseAddr(listen)
		if err == nil {
			li.AddrPort = netip.AddrPortFrom(addr, port)
			goto _default_
		}

		li.Iface = utils.InterfaceByName(listen)
		if li.Iface != nil {
			ip := utils.InterfaceAddr(li.Iface, false)
			if addr, ok := netip.AddrFromSlice(ip.IP); ok {
				li.AddrPort = netip.AddrPortFrom(addr, port)
			}
		} else {
			log.Error("this is not a interface name", zap.String("listen", listen))
		}
	}
	if !li.AddrPort.IsValid() {
		li.AddrPort = netip.AddrPortFrom(netip.MustParseAddr("0.0.0.0"), port)
	}
_default_:
	li.IPV6 = !li.AddrPort.Addr().Is4()

	cfg.Set(reflect.ValueOf(li))
	return
}

func parseBits(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *CfgKey) {
	parse := func(b string) bool {
		r := strings.FieldsFunc(b, func(r rune) bool {
			return r == ' ' || r == '|' || r == '/'
		})

		SupportAudioBits = make([]audio.Bits, 0)
		for _, v := range r {
			var a audio.Bits
			a.FromName(v)
			if a == audio.Bits_NONE {
				log.Error("bits is invalid", zap.String("bits", v), zap.String("key", ck.Key))
				continue
			}
			if utils.SliceContains(SupportAudioBits, a) {
				continue
			}
			SupportAudioBits = append(SupportAudioBits, a)
		}
		return len(SupportAudioBits) > 0
	}
	if k != nil {
		b := k.String()
		if len(b) > 0 && parse(b) {
			return
		}
	}
	parse(def.String())
}

func parseRates(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *CfgKey) {
	parse := func(b string) bool {
		r := strings.FieldsFunc(b, func(r rune) bool {
			return r == ' ' || r == '|' || r == '/'
		})

		SupportAudioRates = make([]audio.Rate, 0)
		for _, v := range r {
			i, err := strconv.ParseInt(v, 0, 32)
			if err != nil {
				log.Error("rate is invalid", zap.String("rate", v), zap.String("key", ck.Key))
				continue
			}
			a := audio.NewAudioRate(int(i))
			if !a.IsValid() {
				log.Error("rate is invalid", zap.String("rate", v), zap.String("key", ck.Key))
				continue
			}
			if utils.SliceContains(SupportAudioRates, a) {
				continue
			}
			SupportAudioRates = append(SupportAudioRates, a)
		}
		return len(SupportAudioRates) > 0
	}
	if k != nil {
		b := k.String()

		if len(b) > 0 && parse(b) {
			return
		}
	}
	parse(def.String())
}

func parsePath(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *CfgKey) {
	path := def.String()
	if k != nil {
		path = k.String()
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error("path not exists", zap.String("path", path), zap.String("key", ck.Key))
		cfg.SetString(def.String())
	} else {
		cfg.SetString(path)
	}
}
func parseTempDir(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *CfgKey) {
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

func Clear(cfg reflect.Value) {
	switch cfg.Kind() {
	case reflect.String:
		cfg.SetString("")
	case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8, reflect.Int64:
		cfg.SetInt(0)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		cfg.SetUint(0)
	case reflect.Bool:
		cfg.SetBool(false)
	case reflect.Struct:
		cfg.Set(reflect.New(cfg.Type()).Elem())
	case reflect.Slice:
		cfg.Set(reflect.MakeSlice(cfg.Type(), 0, 0))
	}
}

type CfgKey struct {
	Cfg  any
	Key  string
	Def  string
	Desc string
	cb   func(log *zap.Logger, cfg reflect.Value, k *ini.Key, def *ini.Key, ck *CfgKey)
}
type CfgSection struct {
	Name string
	Keys []CfgKey
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

	for _, cs := range ConfigStruct {
		section := c.Section(cs.Name)

		for _, ck := range cs.Keys {
			setKey(log, section, &ck)
		}
	}

	if ServerListen.IPV6 {
		MulticastAddress = netip.MustParseAddr("FF02:2C:4D:FF::16")
	} else {
		MulticastAddress = netip.MustParseAddr("239.44.77.16")
	}
	MulticastPort = 4414

	return nil
}

func setKey(log *zap.Logger, section *ini.Section, ck *CfgKey) {

	key, _ := section.GetKey(ck.Key)
	defKey, _ := section.NewKey(ck.Key+"-default", ck.Def)

	isNil := ck.Cfg == nil

	cfgrv := reflect.ValueOf(ck.Cfg)

	if !isNil && cfgrv.Kind() != reflect.Pointer {
		panic(fmt.Errorf("the '%s' must be a pointer", reflect.TypeOf(ck.Cfg).Name()))
	}

	if !isNil {
		cfgrv = cfgrv.Elem()
	}

	Clear(cfgrv)

	if ck.cb != nil {
		ck.cb(log, cfgrv, key, defKey, ck)
		return
	}

	if isNil {
		return
	}

	switch cfgrv.Kind() {
	case reflect.String:
		setReflectString(log, key, defKey, cfgrv, ck)
	case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8, reflect.Int64:
		setReflectInt(log, key, defKey, cfgrv, ck)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		setReflectUInt(log, key, defKey, cfgrv, ck)
	case reflect.Bool:
		setReflectBool(log, key, defKey, cfgrv, ck)
	case reflect.Struct:
		setReflectStruct(log, key, defKey, cfgrv, ck)
	case reflect.Slice:
		setReflectSlice(log, key, defKey, cfgrv, ck)
	}
}

func setReflectString(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *CfgKey) {
	if key == nil {
		cfgrv.SetString(ck.Def)
	} else {
		i := key.String()
		if key == nil || len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.Key), zap.String("val", key.String()))
			i = ck.Def
		}
		cfgrv.SetString(i)
	}
}

func setReflectInt(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *CfgKey) {
	i, err := defKey.Int()
	if err != nil {
		panic(nil)
	}

	if key != nil {
		if len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.Key), zap.String("val", key.String()))
		} else if ki, err := key.Int(); err != nil {
			log.Error("value invalid", zap.String("key", ck.Key), zap.String("val", key.String()))
			i = ki
		}
	}
	cfgrv.SetInt(int64(i))

}
func setReflectUInt(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *CfgKey) {
	i, err := defKey.Uint()
	if err != nil {
		panic(nil)
	}
	if key != nil {
		if len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.Key), zap.String("val", key.String()))
		} else if ki, err := key.Uint(); err != nil {
			log.Error("value invalid", zap.String("key", ck.Key), zap.String("val", key.String()))
			i = ki
		}
	}

	cfgrv.SetUint(uint64(i))

}
func setReflectBool(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *CfgKey) {
	i, err := defKey.Bool()
	if err != nil {
		panic(nil)
	}
	if key != nil {
		if len(key.String()) == 0 {
			log.Error("value empty", zap.String("key", ck.Key), zap.String("val", key.String()))
		} else if ki, err := key.Bool(); err != nil {
			log.Error("value invalid", zap.String("key", ck.Key), zap.String("val", key.String()))
			i = ki
		}
	}
	cfgrv.SetBool(i)
}
func setReflectStruct(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *CfgKey) {
	var keyV string
	isEmpty := key != nil && len(key.String()) == 0

	if key == nil {
		keyV = defKey.String()
	} else {
		keyV = key.String()
	}
	if isEmpty {
		log.Error("value empty", zap.String("key", ck.Key), zap.String("val", key.String()))
	}

	if cfgrv.Kind() == reflect.Pointer {
		cfgrv = cfgrv.Elem()
	}

	as := cfgrv.Type().AssignableTo

	if as(reflect.TypeOf((*Interface)(nil)).Elem()) {
		port, _ := defKey.Int()
		val := ""
		if key != nil {
			val = key.String()
		}
		parseListen(log, cfgrv, val, uint16(port))
	} else if as(reflect.TypeOf((*net.Interface)(nil)).Elem()) {
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

func setReflectSlice(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *CfgKey) {
	isEmpty := key != nil && len(key.String()) == 0

	if isEmpty {
		log.Error("value empty", zap.String("key", ck.Key), zap.String("val", key.String()))
	}
	elemRT := cfgrv.Type().Elem()
	as := elemRT.AssignableTo

	if as(reflect.PointerTo(reflect.TypeOf((*net.IPNet)(nil)).Elem())) {
		setReflectSliceIPNet(log, key, defKey, cfgrv, ck)
	}
}

func setReflectSliceIPNet(log *zap.Logger, key *ini.Key, defKey *ini.Key, cfgrv reflect.Value, ck *CfgKey) {
	var keyV string
	elemRT := cfgrv.Type().Elem()
	if key == nil {
		keyV = defKey.String()
	} else {
		keyV = key.String()
	}
	if len(keyV) == 0 {
		cfgrv.Set(reflect.Zero(cfgrv.Type()))
		return
	}
	keyList := strings.FieldsFunc(keyV, func(r rune) bool {
		return r == ' ' || r == SEP
	})
	for _, ipstr := range keyList {
		_, ipnet, err := net.ParseCIDR(ipstr)
		if err != nil {
			ip := net.ParseIP(ipstr)
			if ip == nil {
				log.Error("ip invalid", zap.String("key", ck.Key), zap.String("val", key.String()))
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

func configToString(cr reflect.Value) (val string, tn string) {
	tn = cr.Type().Name()
	if s, ok := cr.Interface().(fmt.Stringer); ok {
		return s.String(), tn
	}

	switch cr.Kind() {
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return fmt.Sprintf("%d", cr.Uint()), tn
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return fmt.Sprintf("%d", cr.Int()), tn
	case reflect.String:
		return cr.String(), tn
	case reflect.Bool:
		if cr.Bool() {
			return "true", tn
		} else {
			return "false", tn
		}
	case reflect.Slice:
		strs := make([]string, cr.Len())
		var v, k string
		for i := 0; i < cr.Len(); i++ {
			v, k = configToString(cr.Index(i))
			strs[i] = v
		}
		if cr.Len() == 0 {
			se := cr.Type().Elem()
			if se.Kind() == reflect.Pointer {
				se = se.Elem()
			}
			k = se.Name()
		}
		return strings.Join(strs, string(SEP)), "[]" + k
	case reflect.Pointer:
		return configToString(cr.Elem())
	case reflect.Struct:
		if s, ok := cr.Interface().(*net.Interface); ok {
			return s.Name, "iface"
		}
	}

	return "", tn
}

func ConfigString(cs *CfgSection, ck *CfgKey) (string, string) {
	if ck.Cfg == nil {
		return "", Unknown
	}
	cr := reflect.ValueOf(ck.Cfg).Elem()
	return configToString(cr)
}

func Save(fp io.ReadWriter) error {
	c, err := ini.Load([]byte{})
	if err != nil {
		return err
	}
	for _, cs := range ConfigStruct {
		s := c.Section(cs.Name)
		for _, ck := range cs.Keys {
			val, t := ConfigString(&cs, &ck)
			if t == Unknown {
				continue
			}
			s.NewKey(ck.Key, val)
		}
	}

	_, err = c.WriteTo(fp)
	return err
}
