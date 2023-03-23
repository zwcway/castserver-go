package decoder

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
	fs := line.Input.FileStreamer()

	if fs == nil {
		fs = ffmpeg.New(line.Output)
		line.ApplyInput(fs)

		fs.SetFormatChangedHandler(onFileOpened)
	}

	return fs
}

func findLineByFileStreamer(s stream.Streamer) *speaker.Line {
	for _, l := range speaker.LineList() {
		if l.Input.FileStreamer() == s {
			return l
		}
	}
	return nil
}

func onFileOpened(s stream.SourceStreamer, inFormat, outFormat audio.Format) {
	// bufSize := outFormat.SampleBits.Size() * int(inFormat.SampleRate.ToInt()) * 10 / 1000 * inFormat.Layout.Count

	l := findLineByFileStreamer(s)
	if l == nil {
		return
	}

	// pl.Lock()
	// pl.SetBuffer(stream.NewSamples(bufSize, outFormat))
	// pl.Unlock()

	// 通知输入格式
	websockets.BroadcastLineInputEvent(l)
}
