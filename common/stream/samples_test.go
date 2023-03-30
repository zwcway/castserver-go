package stream

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zwcway/castserver-go/common/audio"
)

func TestNewSamplesDuration(t *testing.T) {

	t.Run("sample 1ms", func(t *testing.T) {
		format := audio.Format{
			Sample: audio.Sample{
				Rate: audio.AudioRate_44100,
				Bits: audio.Bits_DEFAULT,
			},
			Layout: audio.Layout10,
		}
		samples := NewSamplesDuration(time.Millisecond, format)

		assert.Equal(t, samples.autoSize, false)
		assert.Equal(t, samples.RequestNbSamples, 44)
		assert.Equal(t, len(samples.buffer), format.SamplesSize(44))
		assert.Equal(t, len(samples.RawData), len(samples.Data), 1)
		assert.Equal(t, samples.fmt, samples.Format, format)
		assert.Equal(t, samples.LastErr, nil)
		assert.Equal(t, samples.LastNbSamples, 0)
	})
}
