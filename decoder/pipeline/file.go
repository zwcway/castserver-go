package pipeline

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/castserver-go/decoder/ffmpeg"
	"github.com/zwcway/castserver-go/decoder/streamer"
)

var fileStream decoder.FileStreamer

func AddFileStreamer() decoder.FileStreamer {
	fileStream = ffmpeg.New(onFileOpened)

	return fileStream
}

func onFileOpened(format *audio.Format) {
	bufSize := format.SampleBits.Size() * int(format.SampleRate.ToInt()) * 30 / 1000 * format.Layout.Count

	globalPipeLine.Prepend(fileStream.(decoder.Element))
	globalPipeLine.format = format
	globalPipeLine.buffer = decoder.NewSamples(bufSize, format)

	// 通知输入格式
	speaker.DefaultLine().Input = format
}

func SetOutputFormat(format *audio.Format) {
	if fileStream == nil {
		globalPipeLine.Add(streamer.ResampleSet(format))
		return
	}
	fileStream.SetFormat(format)
}
