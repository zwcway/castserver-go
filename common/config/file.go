package config

import (
	"fmt"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
	"gorm.io/gorm"

	"github.com/zwcway/castserver-go/common/audio"

	"github.com/go-ini/ini"
)

const (
	SEP     = '|'
	Unknown = "unknown"
)

func parseListen(cfg reflect.Value, listen string, port uint16) {
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
			log.Error("this is not a interface name", lg.String("listen", listen))
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

func parseBits(cfg reflect.Value, k *ini.Key, ck *CfgKey) {
	parse := func(b string) []audio.Bits {
		r := strings.FieldsFunc(b, func(r rune) bool {
			return r == ' ' || r == '|' || r == '/'
		})

		bits := make([]audio.Bits, 0)
		for _, v := range r {
			var a audio.Bits
			a.FromName(v)
			if a == audio.Bits_NONE {
				log.Error("bits is invalid", lg.String("bits", v), lg.String("key", ck.Key))
				continue
			}
			if utils.SliceContains(bits, a) >= 0 {
				continue
			}
			bits = append(bits, a)
		}
		return bits
	}
	if k != nil {
		b := k.String()
		if len(b) > 0 {
			if bits := parse(b); len(bits) > 0 {
				cfg.Set(reflect.ValueOf(bits))
			}
			return
		}
	}
}

func parseRates(cfg reflect.Value, k *ini.Key, ck *CfgKey) {
	parse := func(b string) []audio.Rate {
		r := strings.FieldsFunc(b, func(r rune) bool {
			return r == ' ' || r == '|' || r == '/'
		})

		rates := make([]audio.Rate, 0)
		for _, v := range r {
			i, err := strconv.ParseInt(v, 0, 32)
			if err != nil {
				log.Error("rate is invalid", lg.String("rate", v), lg.String("key", ck.Key))
				continue
			}
			a := audio.NewAudioRate(int(i))
			if !a.IsValid() {
				log.Error("rate is invalid", lg.String("rate", v), lg.String("key", ck.Key))
				continue
			}
			if utils.SliceContains(rates, a) >= 0 {
				continue
			}
			rates = append(rates, a)
		}
		return rates
	}
	if k != nil {
		b := k.String()

		if len(b) > 0 {
			if rates := parse(b); len(rates) > 0 {
				cfg.Set(reflect.ValueOf(rates))
			}
			return
		}
	}
}

func parsePath(cfg reflect.Value, k *ini.Key, ck *CfgKey) {
	if k == nil {
		return
	}
	path := k.String()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error("path not exists", lg.String("path", path), lg.String("key", ck.Key))
	} else {
		cfg.SetString(path)
	}
}

func parseTempDir(cfg reflect.Value, k *ini.Key, ck *CfgKey) {
	path := ""
	if k != nil {
		path = k.String()
	}

	if path == "" {
		path = filepath.Join(os.TempDir(), APPNAME)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModeTemporary)
		if err != nil {
			log.Error("can not create temp dir ", lg.String("dir", path))
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
	Desc string
	cb   func(cfg reflect.Value, k *ini.Key, ck *CfgKey)
}
type CfgSection struct {
	Name string
	Keys []CfgKey
}

func FromFile(ctx utils.Context, file string) error {
	data, err := os.ReadFile(file)

	if err != nil {
		return err
	}

	ConfigFile = file

	return FromContent(ctx, data)
}

func FromContent(ctx utils.Context, data []byte) error {
	initLogger(ctx)

	c, err := ini.Load(data)
	if err != nil {
		return err
	}

	for _, cs := range ConfigStruct {
		section := c.Section(cs.Name)

		for _, ck := range cs.Keys {
			ck.setKey(section)
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
		if s, ok := cr.Interface().(*gorm.DB); ok {
			return getDSN(s), "db"
		}
		return configToString(cr.Elem())
	case reflect.Struct:
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
