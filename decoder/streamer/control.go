package streamer

import (
	"github.com/zwcway/castserver-go/decoder"
)

type control struct {
	pause bool
}

func (c *control) Name() string {
	return "Controller"
}

func (c *control) Type() decoder.ElementType {
	return decoder.ET_OneSample
}

func (c *control) Stream(samples *decoder.Samples) {}

func (c *control) Sample(sample *float64, ch int, n int) {
	if c.pause {
		*sample = 0
	}
}

var controlStreamer = &control{}

func ControlStreamer() decoder.Streamer {
	return controlStreamer
}

func ControlIsPaused() bool {
	return controlStreamer.pause
}

func ControlPause() {
	controlStreamer.pause = true
}

func ControlUnPause() {
	controlStreamer.pause = false
}

func ControlTogglePause() {
	controlStreamer.pause = !controlStreamer.pause
}
