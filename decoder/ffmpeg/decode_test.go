package ffmpeg

import (
	"bytes"
	"testing"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/playlist"
	"github.com/zwcway/castserver-go/common/stream"
)

func TestAVFormatContext_Stream(t *testing.T) {

	t.Run("decode mp3 44100/fltp/stereo", func(t *testing.T) {
		f := audio.Format{
			SampleRate: audio.AudioRate_44100,
			SampleBits: audio.Bits_DEFAULT,
			Layout:     audio.ChannelLayoutMono,
		}
		fs := New(f)
		defer fs.Close()

		err := fs.OpenFile("./test/test_44100_fltp_stereo.mp3")
		if err != nil {
			t.Errorf("open error %v", err)
			return
		}
		samples := stream.NewSamples(10, f)
		fs.SetPause(false)
		fs.Stream(samples)

		if samples.LastNbSamples != 10 {
			t.Errorf("decoded length error %d", samples.LastNbSamples)
			return
		}

		want := []byte{
			0x00, 0x00, 0x00, 0x40, 0x9F, 0x1E, 0x77, 0xBD, 0x00, 0x00, 0x00, 0x80, 0xF5, 0x03, 0x7C, 0xBD,
			0x00, 0x00, 0x00, 0x80, 0x73, 0x4D, 0x80, 0xBD, 0x00, 0x00, 0x00, 0x00, 0x2D, 0x68, 0x82, 0xBD,
			0x00, 0x00, 0x00, 0x80, 0xA5, 0x32, 0x84, 0xBD, 0x00, 0x00, 0x00, 0xC0, 0x5E, 0x6F, 0x85, 0xBD,
			0x00, 0x00, 0x00, 0xE0, 0x12, 0xC6, 0x85, 0xBD, 0x00, 0x00, 0x00, 0xC0, 0x24, 0xD1, 0x84, 0xBD,
			0x00, 0x00, 0x00, 0x80, 0x84, 0x39, 0x82, 0xBD, 0x00, 0x00, 0x00, 0xC0, 0x3E, 0xA1, 0x7B, 0xBD,
		}
		if !bytes.Equal(samples.RawData[0], want) {
			t.Errorf("reader.Read() = \n%v\n, want \n%v\n", samples.RawData[0], want)
			return
		}
	})

}

func TestAudioInfo(t *testing.T) {

	t.Run("get empty audio info", func(t *testing.T) {
		ai := playlist.AudioInfo{}
		if err := AudioInfo("test/test_44100_fltp_stereo.mp3", &ai); err != nil {
			t.Errorf("AudioInfo() error = %v", err)
		}
		t.Logf("%s", ai.Url)
		t.Logf("%s", ai.Format.String())
		t.Logf("%v/%v", ai.Position, ai.Duration)
		t.Logf("%s", ai.Title)
		t.Logf("%s", ai.Artist)
	})
}
