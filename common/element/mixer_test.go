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
	for i := 0; i < 8; i++ {
		samples.Data[0][i] = float64(m.i)
		m.i++
	}
	samples.LastNbSamples = 8
}
func (m mixer1) Close() error { return nil }
func (m mixer1) AudioFormat() audio.Format {
	return audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
		Layout: audio.Layout10,
	}
}
func (m mixer1) SetOutFormat(f audio.Format) error { return nil }
func (m mixer1) CanRemove() bool                   { return false }
func (m mixer1) IsPlaying() bool                   { return true }

func TestMixer(t *testing.T) {
	t.Parallel()

	format := audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
		Layout: audio.Layout10,
	}
	result := []float64{0, 2, 4, 6, 8, 10, 12, 14}
	t.Run("mix size 512 for 8", func(t *testing.T) {
		mixer := NewMixer(&mixer1{}, &mixer1{})
		samples := stream.NewSamples(8, format)
		mixer.Stream(samples)
		if !slices.Equal(samples.Data[0][:8], result) {
			t.Errorf("mix same error = \n%v\n, want \n%v\n", samples.Data[0], result)
		}
	})
	t.Run("mix size 6 for 8", func(t *testing.T) {
		mixer := NewMixer(&mixer1{}, &mixer1{}).(*Mixer)
		samples := stream.NewSamples(8, format)
		mixer.Stream(samples)
		if !slices.Equal(samples.Data[0][:8], result) {
			t.Errorf("mix same error = \n%v\n, want \n%v\n", samples.Data[0], result)
		}
	})
	t.Run("mix size 5 for 8", func(t *testing.T) {
		mixer := NewMixer(&mixer1{}, &mixer1{}).(*Mixer)
		samples := stream.NewSamples(8, format)
		mixer.Stream(samples)
		if !slices.Equal(samples.Data[0][:8], result) {
			t.Errorf("mix same error = \n%v\n, want \n%v\n", samples.Data[0], result)
		}
	})
}
