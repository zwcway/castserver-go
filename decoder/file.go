package decoder

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/ffmpeg"
)

func FileStreamer(uuid string) stream.FileStreamer {
	l := speaker.FindLineByUUID(uuid)
	return FileStreamerFromLine(l)
}

func FileStreamerFromLine(line *speaker.Line) stream.FileStreamer {
	fs := line.Input.FileStreamer()

	if fs == nil {
		fs = ffmpeg.New(audio.InternalFormat())
		line.ApplyInput(fs)
	}

	return fs
}
