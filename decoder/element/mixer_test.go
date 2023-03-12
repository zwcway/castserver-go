package element

import (
	"testing"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
	"golang.org/x/exp/slices"
)

type mixer1 struct {
	i int
}

func (m *mixer1) Stream(samples *stream.Samples) {
	for i := 0; i < samples.NbSamples; i++ {
		samples.Data[0][i] = float64(m.i)
		m.i++
	}
	samples.LastNbSamples = samples.NbSamples
}

func TestMixer(t *testing.T) {
	format := audio.Format{
		SampleRate: audio.AudioRate_44100,
		Layout:     audio.ChannelLayout10,
		SampleBits: audio.Bits_U8,
	}
	result := []float64{0, 2, 4, 6, 8, 10, 12, 14}
	t.Run("mix size 512 for 8", func(t *testing.T) {
		mixer := NewMixer(&mixer1{}, &mixer1{})
		samples := stream.NewSamples(8, format)
		mixer.Stream(samples)
		if !slices.Equal(samples.Data[0], result) {
			t.Error("mix same error")
		}
	})
	t.Run("mix size 6 for 8", func(t *testing.T) {
		mixer := NewMixer(&mixer1{}, &mixer1{}).(*Mixer)
		mixer.bufSize = 6
		samples := stream.NewSamples(8, format)
		mixer.Stream(samples)
		if !slices.Equal(samples.Data[0], result) {
			t.Error("mix same error")
		}
	})
	t.Run("mix size 5 for 8", func(t *testing.T) {
		mixer := NewMixer(&mixer1{}, &mixer1{}).(*Mixer)
		mixer.bufSize = 5
		samples := stream.NewSamples(8, format)
		mixer.Stream(samples)
		if !slices.Equal(samples.Data[0], result) {
			t.Error("mix same error")
		}
	})
}
