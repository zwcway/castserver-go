package pusher

import (
	"time"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
)

type lineTimer struct {
	ticker *time.Ticker
}

var (
	lineList        map[*speaker.Line]lineTimer
	timerQuitSignal chan struct{}
)

func TimerAddLine(line *speaker.Line) {
	TimerRemoveLine(line)
	LineFormatChanged(line)
	go linePushTimerRoutine(line)
}

func LineFormatChanged(line *speaker.Line) {
	pl := line.Input.PipeLine
	if pl == nil || pl.Buffer() == nil {
		return
	}
	pl.SetBuffer(stream.NewSamples(2048, line.Output))

	nbSamples := pl.Buffer().RequestNbSamples

	rate := time.Duration(line.Output.Rate.ToInt())
	t := time.Duration(nbSamples) * time.Second / rate

	lineList[line] = lineTimer{
		ticker: time.NewTicker(t),
	}
}

func TimerRemoveLine(line *speaker.Line) {
	lt, ok := lineList[line]
	if !ok {
		return
	}
	if lt.ticker != nil {
		timerQuitSignal <- struct{}{}
		lt.ticker.Stop()
		lt.ticker = nil
	}
	delete(lineList, line)
}

func linePushTimerRoutine(line *speaker.Line) {
	defer TimerRemoveLine(line)

	for {
		lt, ok := lineList[line]
		if !ok {
			return
		}

		select {
		case <-context.Done():
			return
		case <-timerQuitSignal:
			return
		case <-lt.ticker.C:
		}
		// todo 高精度，所有设备播放时以该时钟为基准
		if currentTrigger != trigger_timer {
			continue
		}

		line.Input.PipeLine.Stream(nil)
	}
}
