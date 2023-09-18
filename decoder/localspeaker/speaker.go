package localspeaker

import (
	"sync"
	"time"

	oto "github.com/hajimehoshi/oto/v2"
	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder"
)

var (
	mu           sync.Mutex
	context      *oto.Context
	player       oto.Player
	sampleReader *reader

	slient bool = false

	mixer stream.MixerElement

	cost time.Duration

	format = audio.Format{
		Sample: audio.Sample{
			Rate: audio.AudioRate_44100,
			Bits: audio.Bits_S16LE,
		},
		Layout: audio.LayoutStereo,
	}
)

func Init() error {
	mu.Lock()
	defer mu.Unlock()
	if player != nil {
		return nil
	}

	var err error
	bufSize := format.AllSamplesSize(1024)

	if sampleReader == nil {
		resample := decoder.NewResample(format)
		sampleReader = &reader{resample: resample}
		resample.On()
	}
	if mixer == nil {
		mixer = element.NewMixer()
	}
	var readyChan chan struct{}

	context, readyChan, err = oto.NewContext(format.Rate.ToInt(), int(format.Count), 2)
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
	speaker.BusLineCreated.Register(AddLine)
	speaker.BusLineDeleted.Register(RemoveLine)
	stream.BusMixerFormatChanged.Register(mixer, func(m stream.MixerElement, format *audio.Format, channelIndex audio.ChannelIndex) error {
		if sampleReader.samples == nil {
			sampleReader.samples = stream.NewSamplesDuration(config.AudioBuferMSDuration, *format)
		} else {
			sampleReader.samples.ResizeDuration(config.AudioBuferMSDuration, *format)
		}
		sampleReader.samples.SetChannelIndex(channelIndex)
		return nil
	})
}

func AddLine(line *speaker.Line) error {
	ss := line.Input.PipeLine.(stream.SourceStreamer)
	if !mixer.Has(ss) {
		mixer.Add(ss)
	}
	return nil
}

func RemoveLine(line, dst *speaker.Line) error {
	ss := line.Input.PipeLine.(stream.SourceStreamer)
	if mixer.Has(ss) {
		mixer.Del(ss)
	}
	return nil
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
	mu.Lock()
	defer mu.Unlock()

	if player == nil || mixer == nil {
		return
	}
	player.Play()

	bus.Dispatch("localspeaker playing")
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
	if mixer != nil {
		mixer.Close()
		mixer = nil
	}

	return nil
}

func Slient(s bool) {
	slient = s
	bus.Dispatch("localspeaker slient", s)
}

func Cost() time.Duration {
	return cost
}
