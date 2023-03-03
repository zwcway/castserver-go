package pipeline

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/castserver-go/decoder/ffmpeg"
	"github.com/zwcway/castserver-go/web/websockets"
)

func FileStreamer(uuid string) decoder.FileStreamer {
	p := FromUUID(uuid)

	fs := p.eleMixer.HasFileStreamer()
	if fs == nil {
		fs = ffmpeg.New(onFileOpened)
		p.eleMixer.Add(fs)
	}

	return fs
}

func findPL(stream decoder.FileStreamer) *PipeLine {
	for _, p := range pipeLineList {
		if p.eleMixer.Has(stream) {
			return p
		}
	}
	return nil
}

func onFileOpened(stream decoder.FileStreamer, format *audio.Format) {
	bufSize := format.SampleBits.Size() * int(format.SampleRate.ToInt()) * 30 / 1000 * format.Layout.Count

	pl := findPL(stream)

	pl.locker.Lock()
	pl.buffer = decoder.NewSamples(bufSize, format)
	pl.locker.Unlock()

	// 通知输入格式
	pl.line.Input = format
	websockets.BroadcastLineEvent(pl.line, websockets.Event_Line_Edited)
}
