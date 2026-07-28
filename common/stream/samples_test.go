package stream

import (
	"fmt"
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

		assert.NotNil(t, samples)
		assert.Equal(t, samples.autoSize, false)
		assert.Equal(t, samples.RequestNbSamples, 44)
		assert.Equal(t, len(samples.buffer), format.SamplesSize(44))
		assert.Equal(t, len(samples.RawData), len(samples.Data), 1)
		assert.Equal(t, samples.fmt, samples.Format, format)
		assert.Equal(t, samples.LastErr, nil)
		assert.Equal(t, samples.LastNbSamples, 0)
	})
}

// Unit tests
// TestNewSamples tests the creation of a new Samples instance
func TestNewSamples(t *testing.T) {
	format := audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
		Layout: audio.Layout20,
	}
	samples := NewSamples(16, format)

	assert.NotNil(t, samples)
	assert.Equal(t, samples.Format, format)
	assert.Equal(t, samples.RequestNbSamples, 16)

	zeroSlice := make([]byte, format.AllSamplesSize(16))
	assert.Equal(t, samples.buffer, zeroSlice)
}

// TestPackedBytes tests conversion from planar to packed format
func TestPackedBytes(t *testing.T) {
	format := audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
	}
	samples := NewSamples(10, format)
	assert.NotNil(t, samples)

	// Fill with test data
	if len(samples.RawData) > 0 && len(samples.RawData[0]) > 0 {
		samples.RawData[0][0] = 0x01
		samples.LastNbSamples = 1
	}

	packed := samples.PackedBytes()

	assert.NotNil(t, packed)
	assert.Greater(t, len(packed), 0)
}

// TestCopyToChannels tests copying data between Samples objects
func TestCopyTo(t *testing.T) {
	format := audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
	}
	src := NewSamples(10, format)
	dst := NewSamples(10, format)

	if src == nil || dst == nil {
		t.Fatalf("NewSamples returned nil")
	}

	if len(src.RawData) > 0 && len(src.RawData[0]) > 0 {
		src.RawData[0][0] = 0x42
		src.LastNbSamples = 1

		copied := src.CopyTo(dst, 0, 0)
		if copied != 1 {
			t.Fatalf("CopyToChannels: expected 1 byte copied, got %d", copied)
		}

		if dst.RawData[0][0] != 0x42 {
			t.Fatalf("CopyToChannels: data not copied correctly, expected 0x42, got 0x%02x", dst.RawData[0][0])
		}
	}
}

// TestResize tests resizing Samples buffer
func TestResize(t *testing.T) {
	format := audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
	}
	samples := NewSamples(10, format)
	if samples == nil {
		t.Fatalf("NewSamples returned nil")
	}

	// oldSize := samples.SamplesCount()
	samples.Resize(20, format)
	newSize := samples.SamplesCount()

	if newSize < 20 {
		t.Fatalf("Resize failed: expected at least 20 samples, got %d", newSize)
	}
}

// TestChannelSamples tests extracting a single channel
func TestChannelSamples(t *testing.T) {
	format := audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
	}
	samples := NewSamples(10, format)
	if samples == nil {
		t.Fatalf("NewSamples returned nil")
	}

	chSamples := samples.ChannelSamples(audio.ChannelFrontLeft)
	if chSamples == nil {
		t.Fatalf("ChannelSamples returned nil")
	}
	if chSamples.Format.Count != 1 {
		t.Fatalf("ChannelSamples: expected 1 channel, got %d", chSamples.Format.Count)
	}
}

// TestWrapError tests error wrapping
func TestWrapError(t *testing.T) {
	format := audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_DEFAULT,
		},
	}
	samples := NewSamples(10, format)
	if samples == nil {
		t.Fatalf("NewSamples returned nil")
	}

	err1 := fmt.Errorf("error 1")
	samples.WrapError(err1)
	if samples.LastErr == nil {
		t.Fatalf("WrapError: error not set")
	}

	err2 := fmt.Errorf("error 2")
	samples.WrapError(err2)
	if samples.LastErr == nil {
		t.Fatalf("WrapError: error not wrapped")
	}
}
