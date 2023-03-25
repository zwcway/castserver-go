package localspeaker

import (
	"sync"
	"time"

	oto "github.com/hajimehoshi/oto/v2"
	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/utils"
)

var (
	mu           sync.Mutex
	context      *oto.Context
	player       oto.Player
	sampleReader *reader

	slient bool = false

	sampleSize int = 2048
	mixer      stream.MixerElement

	cost  time.Duration
	lines []*speaker.Line
)

func Init() error {
	mu.Lock()
	defer mu.Unlock()
	if player != nil {
		return nil
	}

	format := audio.Format{
		SampleRate: audio.AudioRate_44100,
		Layout:     audio.ChannelLayoutStereo,
		SampleBits: audio.Bits_S16LE,
	}

	var err error
	sampleSize = config.AudioBuferSize
	bufSize := format.AllSamplesSize(1024)

	if sampleReader == nil {
		resample := element.NewResample(format)
		sampleReader = &reader{resample: resample}
		resample.On()
	}
	if mixer == nil {
		mixer = element.NewEmptyMixer()
	}
	var readyChan chan struct{}

	context, readyChan, err = oto.NewContext(format.SampleRate.ToInt(), format.Layout.Count, 2)
	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker")
	}
	<-readyChan

	player = context.NewPlayer(sampleReader)
	if player == nil {
		return errors.New("create player failed")
	}
	player.(oto.BufferSizeSetter).SetBufferSize(bufSize)

	registerEventBus()

	return nil
}

func registerEventBus() {
	bus.Register("line created", func(a ...any) error {
		AddLine(a[0].(*speaker.Line))
		return nil
	})
	bus.Register("line deleted", func(a ...any) error {
		RemoveLine(a[0].(*speaker.Line))
		return nil
	})
}

func AddLine(line *speaker.Line) {
	if mixer == nil {
		return
	}
	for _, l := range lines {
		if l == line {
			return
		}
	}
	format := audio.DefaultFormat()

	mixer.Add(line.Input.PipeLine)
	lines = append(lines, line)

	// TODO 每个 Line 很有可能格式不一致
	if sampleReader.samples == nil || sampleReader.samples.Format.Size() < format.Size() {
		sampleReader.samples = stream.NewSamples(sampleSize, format)
	}
}

func RemoveLine(line *speaker.Line) {
	if mixer == nil {
		return
	}
	for i, l := range lines {
		if l == line {
			mixer.Del(line.Input.PipeLine)
			utils.SliceQuickRemove(&lines, i)
			return
		}
	}
}

func IsOpened() bool {
	return player != nil
}

func IsPlaying() bool {
	if player == nil {
		return false
	}
	return player.IsPlaying()
}

func Play() {
	if player == nil || mixer == nil {
		return
	}
	player.Play()

	bus.Trigger("localspeaker playing")
}

func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if player != nil {
		if err := player.Close(); err != nil {
			return err
		}
		player = nil
	}

	return nil
}

func Slient(s bool) {
	slient = s
}

func Cost() time.Duration {
	return cost
}
