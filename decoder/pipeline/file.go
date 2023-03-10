package pipeline

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/ffmpeg"
	"github.com/zwcway/castserver-go/web/websockets"
)

func FileStreamer(uuid string) stream.FileStreamer {
	l := speaker.FindLineByUUID(uuid)
	return FileStreamerFromLine(l)
}

func FileStreamerFromLine(line *speaker.Line) stream.FileStreamer {
	fs := line.Mixer.FileStreamer()

	if fs == nil {
		fs = ffmpeg.New(onFileOpened)
		line.Mixer.AddFileStreamer(fs)
	}

	return fs
}

func findLineByFileStreamer(s stream.FileStreamer) *speaker.Line {
	for _, l := range speaker.LineList() {
		if l.Mixer.Has(s) {
			return l
		}
	}
	return nil
}

func onFileOpened(s stream.FileStreamer, format *audio.Format) {
	bufSize := format.SampleBits.Size() * int(format.SampleRate.ToInt()) * 10 / 1000 * format.Layout.Count

	l := findLineByFileStreamer(s)
	if l == nil {
		return
	}
	pl := l.Input.PipeLine

	pl.Lock()
	pl.SetBuffer(stream.NewSamples(bufSize, format))
	pl.Unlock()

	// 通知输入格式
	l.Input.FromFileStreamer(s)

	websockets.BroadcastLineInputEvent(l)
}
