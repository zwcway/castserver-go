package local

import (
	"fmt"

	"github.com/ebitengine/oto/v3"
	"github.com/zwcway/castserver-go/common/audio"
)

func otoFormat(f audio.Format) oto.Format {
	switch f.Bits {
	case audio.Bits_32LEF:
		return oto.FormatFloat32LE
	case audio.Bits_S16LE:
		return oto.FormatSignedInt16LE
	case audio.Bits_U8:
		return oto.FormatUnsignedInt8
	default:
		panic(fmt.Sprintf("unexpected audio.Bits: %#v", f.Bits))
	}
}
