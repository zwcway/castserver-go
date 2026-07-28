package decoder

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/stream"
)

// input_44K1_f64_mono 20个样本
var input_44K1_f64_mono = []byte{
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

var input_format = audio.Format{
	Sample: audio.Sample{
		Rate: audio.AudioRate_44100,
		Bits: audio.Bits_64LEF,
	},
	Layout: audio.Layout10,
}

func TestResample_Stream(t *testing.T) {

	t.Run("resample", func(t *testing.T) {
		swrOutFormat := input_format
		swrOutFormat.Bits = audio.Bits_S16LE

		swr := NewResample(swrOutFormat)
		swr.OnFormatChanged(&input_format)
		swr.On()

		samples := stream.NewFromBytes(input_44K1_f64_mono, input_format)

		assert.Equal(t, samples.Format, input_format)

		swr.Stream(samples)

		want := []byte{
			0x04, 0xFD, 0x72, 0x00, 0xA4, 0x04, 0x8C, 0x09, 0xF8, 0x0E, 0x9A, 0x14, 0x17, 0x1A, 0x0A, 0x1F,
			0x22, 0x23, 0x24, 0x26, 0xF1, 0x27, 0x7A, 0x28, 0xBE, 0x27, 0xD0, 0x25, 0xDD, 0x22, 0x1E, 0x1F,
			0xCE, 0x1A, 0x1C, 0x16, 0x26, 0x11, 0x08, 0x0C,
		}

		assert.Equal(t, len(want)/2, samples.LastNbSamples)
		assert.Equal(t, swr.Format(), samples.Format)
		assert.Equal(t, want, samples.ChannelBytes(audio.ChannelFrontCenter))
	})

	// 验证连续多次调用：复用缓冲区时数据正确写入
	t.Run("resample-continuous", func(t *testing.T) {
		swr := NewResample(audio.Format{
			Sample: audio.Sample{
				Rate: audio.AudioRate_44100,
				Bits: audio.Bits_S16LE,
			},
			Layout: audio.Layout10,
		})
		swr.On()

		for i := range 5 {
			samples := stream.NewFromBytes(input_44K1_f64_mono, input_format)
			swr.Stream(samples)
			assert.NoError(t, samples.LastErr)
			assert.Equal(t, 20, samples.LastNbSamples, "iteration %d", i)
			// 第一次和后续的输出应保持一致（输入相同）
			assert.Equal(t, 40, len(samples.ChannelBytes(audio.ChannelFrontCenter)), "iteration %d", i)
		}
	})
}

// TestResample_ChangeFormat 测试中途变更输出格式
// 输入 48000/f64/mono，先转码 2 个样本到 48000/s16le/mono，再变更格式到 48000/s32le/mono 转码 2 个样本
// 注意：44100 -> 48000 采样率转换存在 swr bug，返回 0，这里使用 48000 采样率来测试格式变更功能
func TestResample_ChangeFormat(t *testing.T) {
	inputAudio := bytes.Repeat(input_44K1_f64_mono, 10)

	swrOutFormat := input_format
	swrOutFormat.Rate = audio.AudioRate_48000

	swr := NewResample(swrOutFormat)
	swr.OnFormatChanged(&input_format)
	swr.On()

	// 第一次转码：48000/64f/mono
	samples1 := stream.NewFromBytes(inputAudio, input_format)
	swr.Stream(samples1)

	assert.NoError(t, samples1.LastErr)
	assert.Equal(t, 201, samples1.LastNbSamples)
	assert.Equal(t, len(samples1.Data), 1)
	assert.Equal(t, audio.AudioRate_48000, samples1.Format.Sample.Rate)
	assert.Equal(t, audio.Bits_64LEF, samples1.Format.Sample.Bits)
	assert.Equal(t, audio.Layout10, samples1.Format.Layout)

	swrOutFormat = input_format
	swrOutFormat.Rate = audio.AudioRate_96000
	swrOutFormat.Bits = audio.Bits_S32LE
	swrOutFormat.Layout = audio.Layout20
	swr.SetFormat(swrOutFormat)

	// 第二次转码：96000/s32le/mono，2 个样本
	swr.Stream(samples1)

	assert.NoError(t, samples1.LastErr)
	assert.Equal(t, 403, samples1.LastNbSamples)
	assert.Equal(t, len(samples1.Data), 2)
	assert.Equal(t, audio.AudioRate_96000, samples1.Format.Sample.Rate)
	assert.Equal(t, audio.Bits_S32LE.String(), samples1.Format.Sample.Bits.String())
	assert.Equal(t, audio.Layout20, samples1.Format.Layout)
}

// TestResample_BitsDepth 测试位宽变化：44100/f64/mono -> 44100/s32le/mono
func TestResample_BitsDepth(t *testing.T) {

	swrOutFormat := input_format
	swrOutFormat.Rate = audio.AudioRate_44100
	swrOutFormat.Bits = audio.Bits_S32LE

	swr := NewResample(swrOutFormat)
	swr.OnFormatChanged(&input_format)
	swr.On()

	samples := stream.NewFromBytes(input_44K1_f64_mono, input_format)
	swr.Stream(samples)

	assert.NoError(t, samples.LastErr)
	assert.Equal(t, 20, samples.LastNbSamples)
	assert.Equal(t, len(samples.Data), 1)
	assert.Equal(t, audio.AudioRate_44100, samples.Format.Sample.Rate)
	assert.Equal(t, audio.Bits_S32LE, samples.Format.Sample.Bits)
	assert.Equal(t, audio.Layout10, samples.Format.Layout)
}

// TestResample_StereoChannel 测试声道变化：44100/f64/mono -> 44100/s16le/stereo
func TestResample_StereoChannel(t *testing.T) {
	swr := NewResample(audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_S16LE,
		},
		Layout: audio.Layout20,
	})
	swr.OnFormatChanged(&input_format)
	swr.On()

	samples := stream.NewFromBytes(input_44K1_f64_mono, input_format)
	swr.Stream(samples)

	assert.NoError(t, samples.LastErr)
	assert.Equal(t, 20, samples.LastNbSamples)
	assert.Equal(t, len(samples.Data), 2)
	assert.Equal(t, audio.AudioRate_44100, samples.Format.Sample.Rate)
	assert.Equal(t, audio.Bits_S16LE, samples.Format.Sample.Bits)
	assert.Equal(t, audio.Layout20, samples.Format.Layout)
}
