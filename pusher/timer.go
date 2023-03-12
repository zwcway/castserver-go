package pusher

import (
	"time"

	"github.com/zwcway/castserver-go/common/speaker"
)

func InitTimer(line *speaker.Line) {
	CloseTimer(line)

	pl := line.Input.PipeLine
	if pl == nil || pl.Buffer() == nil {
		return
	}
	fs := line.Mixer.FileStreamer()
	if fs == nil {
		return
	}

	format := fs.AudioFormat()
	nbSamples := pl.Buffer().Size

	rate := time.Duration(format.SampleRate.ToInt())
	t := time.Duration(nbSamples) * time.Second / rate

	line.Ticker = time.NewTicker(t)
	go linePushTimerRoutine(line)
}

func CloseTimer(line *speaker.Line) {
	if line.Ticker != nil {
		line.Ticker.Stop()
		line.Ticker = nil
	}
}

func linePushTimerRoutine(line *speaker.Line) {
	defer CloseTimer(line)

	for {
		select {
		case <-context.Done():
			return
		case <-line.Ticker.C:
		}
		// todo 高精度，所有设备播放时以该时钟为基准

		pl := line.Input.PipeLine
		if pl == nil || pl.Buffer() == nil {
			return
		}
		buf := pl.Buffer()

		pl.Stream(buf)

		chs := buf.Format.Layout.Mask.Slice()
		for i, ch := range chs {
			PushToLineChannel(line, ch, buf.ToPacked(i))
		}
	}
}
