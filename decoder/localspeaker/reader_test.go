package localspeaker

import (
	"bytes"
	"os"
	"testing"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/decoder"
)

func Test_reader_Read(t *testing.T) {
	data := []byte{
		0x00, 0x00, 0x00, 0x80, 0xD2, 0xE3, 0x97, 0xBF, 0x00, 0x00, 0x00, 0x00, 0x9A, 0x76, 0x6C, 0x3F,
		0x00, 0x00, 0x00, 0xA0, 0x99, 0x8F, 0xA2, 0x3F, 0x00, 0x00, 0x00, 0xE0, 0xE4, 0x18, 0xB3, 0x3F,
		0x00, 0x00, 0x00, 0x80, 0x93, 0xEF, 0xBD, 0x3F, 0x00, 0x00, 0x00, 0xA0, 0x1D, 0x9A, 0xC4, 0x3F,
		0x00, 0x00, 0x00, 0xC0, 0xAB, 0x16, 0xCA, 0x3F, 0x00, 0x00, 0x00, 0x20, 0x3D, 0x0A, 0xCF, 0x3F,
		0x00, 0x00, 0x00, 0x00, 0xC5, 0x90, 0xD1, 0x3F, 0x00, 0x00, 0x00, 0xE0, 0x11, 0x12, 0xD3, 0x3F,
		0x00, 0x00, 0x00, 0xE0, 0xBB, 0xF8, 0xD3, 0x3F, 0x00, 0x00, 0x00, 0x20, 0x0A, 0x3D, 0xD4, 0x3F,
		0x00, 0x00, 0x00, 0xC0, 0xF5, 0xDE, 0xD3, 0x3F, 0x00, 0x00, 0x00, 0x80, 0x36, 0xE8, 0xD2, 0x3F,
		0x00, 0x00, 0x00, 0xE0, 0x7D, 0x6E, 0xD1, 0x3F, 0x00, 0x00, 0x00, 0xE0, 0x1C, 0x1E, 0xCF, 0x3F,
		0x00, 0x00, 0x00, 0xA0, 0x71, 0xCE, 0xCA, 0x3F, 0x00, 0x00, 0x00, 0xA0, 0xB0, 0x1B, 0xC6, 0x3F,
		0x00, 0x00, 0x00, 0xE0, 0x73, 0x26, 0xC1, 0x3F, 0x00, 0x00, 0x00, 0x60, 0x05, 0x0F, 0xB8, 0x3F,
	}
	mixer = element.NewMixer()
	t.Run("44100/s16le/2=>44100/s16le/2", func(t *testing.T) {
		inFormat := audio.Format{
			Sample: audio.Sample{
				Rate: audio.AudioRate_44100,
				Bits: audio.Bits_64LEF,
			},
			Layout: audio.Layout10,
		}
		pl := pipeline.NewPipeLine(inFormat)
		player := element.NewPlayer()
		player.AddPCMWithChannel(audio.Channel_FRONT_CENTER, inFormat, data)
		pl.Append(player)

		// 输出格式
		format = audio.Format{
			Sample: audio.Sample{
				Rate: audio.AudioRate_44100,
				Bits: audio.Bits_S16LE,
			},
			Layout: audio.Layout10,
		}
		r := reader{
			samples:  stream.NewSamples(512, inFormat),
			resample: decoder.NewResample(format),
		}
		r.resample.On()

		mixer.Clear()
		mixer.Add(pl.(stream.SourceStreamer))

		p := make([]byte, 40)

		gotN, err := r.Read(p)
		if err != nil {
			t.Errorf("reader.Read() error = %v", err)
			return
		}
		if gotN != len(p) {
			t.Errorf("reader.Read() = %v, want %v", gotN, len(p))
		}
		want := []byte{
			0x04, 0xFD, 0x72, 0x00, 0xA4, 0x04, 0x8C, 0x09, 0xF8, 0x0E, 0x9A, 0x14, 0x17, 0x1A, 0x0A, 0x1F,
			0x22, 0x23, 0x24, 0x26, 0xF1, 0x27, 0x7A, 0x28, 0xBE, 0x27, 0xD0, 0x25, 0xDD, 0x22, 0x1E, 0x1F,
			0xCE, 0x1A, 0x1C, 0x16, 0x26, 0x11, 0x08, 0x0C,
		}
		if !bytes.Equal(p, want) {
			t.Errorf("reader.Read() = \n%v\n, want \n%v\n", p, want)
		}
	})

}

func TestMain(m *testing.M) {
	ctx := utils.NewEmptyContext()
	bus.Init(ctx)
	decoder.Module.Init(ctx)
	m.Run()
	os.Exit(0)
}
