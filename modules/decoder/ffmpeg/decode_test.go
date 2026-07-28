package ffmpeg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/playlist"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
)

func TestAVFormatContext_Stream(t *testing.T) {

	t.Run("decode mp3 44100/fltp/mono", func(t *testing.T) {
		f := audio.Format{
			Sample: audio.Sample{
				Rate: audio.AudioRate_44100,
				Bits: audio.Bits_32LEF,
			},
			Layout: audio.LayoutMono,
		}
		fs := New(f).(*AVFormatContext)
		defer fs.Close()

		err := fs.OpenFile("./test/test_44100_fltp_mono.mp3")
		if err != nil {
			t.Errorf("open error %v", err)
			return
		}
		samples := stream.NewSamples(10, f)
		fs.SetPause(false)
		fs.Stream(samples)

		assert.Equal(t, 10, samples.LastNbSamples)

		// ffmpeg -loglevel 0 -nostats -i test/test_44100_fltp_mono.mp3 -af "atrim=start_sample=0:end_sample=20" -f f32le -acodec pcm_f32le  - |  hexdump -v -e '1/1 "0x%02X" ","' -n 40
		want := []byte{
			0xFA, 0xF4, 0xB8, 0xAB, 0xAC, 0x1F, 0xE0, 0xAB, 0x9C, 0x6B, 0x02, 0xAC, 0x68, 0x41, 0x13, 0xAC,
			0x2C, 0x95, 0x21, 0xAC, 0xF6, 0x7A, 0x2B, 0xAC, 0x97, 0x30, 0x2E, 0xAC, 0x26, 0x89, 0x26, 0xAC,
			0x24, 0xCC, 0x11, 0xAC, 0xF6, 0x09, 0xDD, 0xAB,
		}
		assert.Equal(t, want, samples.RawData[0])
	})

}

func TestAudioInfo(t *testing.T) {

	t.Run("get empty audio info", func(t *testing.T) {
		ai := playlist.AudioInfo{}
		if err := AudioInfo("test/test_44100_fltp_mono.mp3", &ai); err != nil {
			t.Errorf("AudioInfo() error = %v", err)
		}
		t.Logf("%s", ai.Url)
		t.Logf("%s", ai.Format.String())
		t.Logf("%v/%v", ai.Position, ai.Duration)
		t.Logf("%s", ai.Title)
		t.Logf("%s", ai.Artist)
	})
}

func TestMain(m *testing.M) {
	bus.Init(utils.NewEmptyContext())
	code := m.Run()
	os.Exit(code)
}
