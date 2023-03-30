package config

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/utils"
)

func (ck *CfgKey) setKey(section *ini.Section) {
	key, _ := section.GetKey(ck.Key)

	isNil := ck.Cfg == nil

	cfgrv := reflect.ValueOf(ck.Cfg)

	if !isNil && cfgrv.Kind() != reflect.Pointer {
		panic(fmt.Errorf("the '%s' must be a pointer", reflect.TypeOf(ck.Cfg).Name()))
	}

	if !isNil {
		cfgrv = cfgrv.Elem()
	}

	// Clear(cfgrv)

	if ck.cb != nil {
		ck.cb(cfgrv, key, ck)
		return
	}

	if isNil {
		return
	}
	if key == nil {
		return
	}
	if len(key.String()) == 0 {
		log.Error("value empty", lg.String("key", ck.Key), lg.String("val", key.String()))
		return
	}
	err := ck.setReflect(key, cfgrv)
	if err != nil {
		log.Error(err.Error(), lg.String("key", ck.Key), lg.String("val", key.String()))
	}
}

func (ck *CfgKey) setReflect(key *ini.Key, cfgrv reflect.Value) error {
	switch cfgrv.Kind() {
	case reflect.String:
		return ck.setReflectString(key, cfgrv)
	case reflect.Int32, reflect.Int, reflect.Int16, reflect.Int8, reflect.Int64:
		return ck.setReflectInt(key, cfgrv)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return ck.setReflectUInt(key, cfgrv)
	case reflect.Bool:
		return ck.setReflectBool(key, cfgrv)
	case reflect.Struct:
		return ck.setReflectStruct(key, cfgrv)
	case reflect.Slice:
		return ck.setReflectSlice(key, cfgrv)
	}
	panic("unsupported value")
}

func (ck *CfgKey) setReflectString(key *ini.Key, cfgrv reflect.Value) error {
	i := key.String()
	cfgrv.SetString(i)
	return nil
}

func (ck *CfgKey) setReflectInt(key *ini.Key, cfgrv reflect.Value) error {
	ki, err := key.Int()
	if err == nil {
		cfgrv.SetInt(int64(ki))
	}
	return err
}

func (ck *CfgKey) setReflectUInt(key *ini.Key, cfgrv reflect.Value) error {
	ki, err := key.Uint()
	if err == nil {
		cfgrv.SetUint(uint64(ki))
	}
	return err
}

func (ck *CfgKey) setReflectBool(key *ini.Key, cfgrv reflect.Value) error {
	ki, err := key.Bool()
	if err == nil {
		cfgrv.SetBool(ki)
	}
	return err
}

func (ck *CfgKey) setReflectStruct(key *ini.Key, cfgrv reflect.Value) error {
	keyV := key.String()

	if cfgrv.Kind() == reflect.Pointer {
		cfgrv = cfgrv.Elem()
	}

	as := cfgrv.Type().AssignableTo

	if as(reflect.TypeOf((*Interface)(nil)).Elem()) {
		ls := cfgrv.Interface().(Interface)
		port := ls.AddrPort.Port()
		parseListen(cfgrv, key.String(), uint16(port))
	} else if as(reflect.TypeOf((*MilliDuration)(nil)).Elem()) {
		ki, err := key.Uint64()
		if err == nil {
			cfgrv.SetUint(ki * uint64(time.Millisecond))
		}
		return err
	} else if as(reflect.TypeOf((*net.Interface)(nil)).Elem()) {
		ifi := utils.InterfaceByName(keyV)
		if ifi == nil {
			return &InvalidError{ck, "this is not a interface name"}
		}
		cfgrv.Set(reflect.ValueOf(ifi))
	}

	return nil
}

func (ck *CfgKey) setReflectSlice(key *ini.Key, cfgrv reflect.Value) error {
	elemRT := cfgrv.Type().Elem()
	as := elemRT.AssignableTo

	if as(reflect.PointerTo(reflect.TypeOf((*net.IPNet)(nil)).Elem())) {
		return ck.setReflectSliceIPNet(key, cfgrv)
	}

	panic("unsupported value type")
}

func (ck *CfgKey) setReflectSliceIPNet(key *ini.Key, cfgrv reflect.Value) error {
	var keyV = key.String()
	elemRT := cfgrv.Type().Elem()

	keyList := strings.FieldsFunc(keyV, func(r rune) bool {
		return r == ' ' || r == SEP
	})
	for _, ipstr := range keyList {
		_, ipnet, err := net.ParseCIDR(ipstr)
		if err != nil {
			ip := net.ParseIP(ipstr)
			if ip == nil {
				return &InvalidError{ck, "ip invalid"}
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

	return nil
}
