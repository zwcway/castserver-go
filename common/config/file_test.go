package config

import (
	"testing"

	"github.com/go-ini/ini"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/utils"
)

func TestSetKey(t *testing.T) {
	t.Parallel()

	initLogger(utils.NewEmptyContext())

	iface := utils.DefaultInterface()

	tests := []struct {
		name    string
		args    string
		test    func() bool
		wantErr bool
	}{
		{"detect listen interface", "[detect]\nlisten: " + iface.Name, func() bool {
			return ServerListen.Iface != nil && ServerListen.Iface.Name == iface.Name
		}, false},
		{"http listen interface", "[http]\nlisten: " + iface.Name, func() bool {
			return HTTPListen.Iface != nil && HTTPListen.Iface.Name == iface.Name
		}, false},
		{"receive listen interface", "[receive]\nlisten: " + iface.Name, func() bool {
			return ReceiveListen.Iface != nil && ReceiveListen.Iface.Name == iface.Name
		}, false},
		{"dlna listen interface", "[dlna]\nlisten: " + iface.Name, func() bool {
			return DLNAListen.Iface != nil && DLNAListen.Iface.Name == iface.Name
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

			ck.setKey(cSec)

			if !tt.test() {
				t.Errorf("setKey() error")
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	type args struct {
		cs *CfgSection
		ck *CfgKey
	}
	type test struct {
		name  string
		args  args
		want  string
		want1 string
	}
	tests := []test{}
	FromContent(utils.NewEmptyContext(), []byte{})

	for i := 0; i < len(ConfigStruct); i++ {
		for k := 0; k < len(ConfigStruct[i].Keys); k++ {
			tests = append(tests, test{
				name: ConfigStruct[i].Name + "." + ConfigStruct[i].Keys[k].Key,
				args: args{&ConfigStruct[i], &ConfigStruct[i].Keys[k]},
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, tp := ConfigString(tt.args.cs, tt.args.ck)
			t.Logf("%s(%s) = %s", tt.name, tp, val)
		})
	}
}
