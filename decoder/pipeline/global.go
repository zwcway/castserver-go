package pipeline

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/streamer"
)

var globalPipeLine = &PipeLine{
	line: speaker.DefaultLine(),
}

func SetFormat(fmt *audio.Format) *PipeLine {
	globalPipeLine.format = fmt

	return globalPipeLine
}

func Global() *PipeLine {
	return globalPipeLine
}

func AddResampleStream() {
	globalPipeLine.Add(streamer.ResampleStreamer())
}
