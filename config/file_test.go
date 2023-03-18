package config

import (
	"net"
	"testing"

	"github.com/go-ini/ini"
	"github.com/zwcway/castserver-go/common/audio"
	"go.uber.org/zap"
)

func TestSetKey(t *testing.T) {
	log := zap.NewNop()

	ifaces, err := net.Interfaces()
	if err != nil || len(ifaces) == 0 {
		t.Error(err)
		return
	}

	tests := []struct {
		name    string
		args    string
		test    func() bool
		wantErr bool
	}{
		{"detect listen interface", "[detect]\nlisten: " + ifaces[0].Name, func() bool {
			return ServerListen.Iface != nil && ServerListen.Iface.Name == ifaces[0].Name
		}, false},
		{"http listen interface", "[http]\nlisten: " + ifaces[0].Name, func() bool {
			return HTTPListen.Iface != nil && HTTPListen.Iface.Name == ifaces[0].Name
		}, false},
		{"receive listen interface", "[receive]\nlisten: " + ifaces[0].Name, func() bool {
			return ReceiveListen.Iface != nil && ReceiveListen.Iface.Name == ifaces[0].Name
		}, false},
		{"dlna listen interface", "[dlna]\nlisten: " + ifaces[0].Name, func() bool {
			return DLNAListen.Iface != nil && DLNAListen.Iface.Name == ifaces[0].Name
		}, false},
		{"detect listen addr", "[detect]\nlisten: 0.0.0.0", func() bool {
			return ServerListen.Iface == nil && ServerListen.AddrPort.String() == "0.0.0.0:4414"
		}, false},
		{"http listen addr", "[http]\nlisten:  0.0.0.0", func() bool {
			return HTTPListen.Iface == nil && HTTPListen.AddrPort.String() == "0.0.0.0:4415"
		}, false},
		{"receive listen addr", "[receive]\nlisten:  0.0.0.0", func() bool {
			return ReceiveListen.Iface == nil && ReceiveListen.AddrPort.String() == "0.0.0.0:4416"
		}, false},
		{"dlna listen addr", "[dlna]\nlisten:  0.0.0.0", func() bool {
			return DLNAListen.Iface == nil && DLNAListen.AddrPort.String() == "0.0.0.0:4416"
		}, false},
		{"detect listen addrport", "[detect]\nlisten: 0.0.0.0:123", func() bool {
			return ServerListen.Iface == nil && ServerListen.AddrPort.String() == "0.0.0.0:123"
		}, false},
		{"http listen addrport", "[http]\nlisten:  0.0.0.0:123", func() bool {
			return HTTPListen.Iface == nil && HTTPListen.AddrPort.String() == "0.0.0.0:123"
		}, false},
		{"receive listen addrport", "[receive]\nlisten:  0.0.0.0:123", func() bool {
			return ReceiveListen.Iface == nil && ReceiveListen.AddrPort.String() == "0.0.0.0:123"
		}, false},
		{"dlna listen addrport", "[dlna]\nlisten:  0.0.0.0:123", func() bool {
			return DLNAListen.Iface == nil && DLNAListen.AddrPort.String() == "0.0.0.0:123"
		}, false},
		{"rates", "[audio]\nsupport rates: 44100|44100", func() bool {
			return len(SupportAudioRates) == 1 && SupportAudioRates[0] == audio.AudioRate_44100
		}, false},
		{"rates unknown", "[audio]\nsupport rates: 44111", func() bool {
			return len(SupportAudioRates) > 0
		}, false},
		{"bits", "[audio]\nsupport bits: 8|16", func() bool {
			return len(SupportAudioBits) == 2 && SupportAudioBits[0] == audio.Bits_U8 && SupportAudioBits[1] == audio.Bits_S16LE
		}, false},
		{"bits unknown", "[audio]\nsupport bits: 4", func() bool {
			return len(SupportAudioBits) > 0
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := ini.Load([]byte(tt.args))
			if (err != nil) != tt.wantErr {
				t.Errorf("FromContent() error = %v, wantErr %v", err, tt.wantErr)
			}
			var section = c.SectionStrings()[1]
			var cSec = c.Section(section)
			var key = cSec.KeyStrings()[0]
			var ck *CfgKey
			for _, cs := range ConfigStruct {
				if cs.Name != section {
					continue
				}
				for _, k := range cs.Keys {
					if k.Key != key {
						continue
					}
					ck = &k
					break
				}
				break
			}
			if ck == nil {
				t.Errorf("setKey() error not found cfgKey %s.%s", section, key)
				return
			}

			setKey(log, cSec, ck)

			if !tt.test() {
				t.Errorf("setKey() error")
			}
		})
	}
}