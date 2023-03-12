package localspeaker

import (
	"sync"
	"time"

	oto "github.com/hajimehoshi/oto/v2"
	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/decoder/element"
	"github.com/zwcway/castserver-go/utils"
)

type PlayCallBackHandler func(*speaker.Line, *stream.Samples)

var (
	mu           sync.Mutex
	context      *oto.Context
	player       oto.Player
	sampleReader *reader

	playCallback PlayCallBackHandler
	slient       bool = false

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

	return nil
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
	pl := line.Input.PipeLine

	mixer.Add(pl)
	lines = append(lines, line)

	// TODO 每个 Line 很有可能格式不一致
	sampleReader.samples = stream.NewSamples(sampleSize, pl.Format())
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

func SetCallback(c PlayCallBackHandler) {
	playCallback = c
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

type reader struct {
	samples  *stream.Samples
	resample stream.ResampleElement
	bufPos   int
	bufSize  int
}

func (r *reader) Read(p []byte) (n int, err error) {
	var (
		samples = r.samples
		i       int
		t       = time.Now()
	)
	defer func() {
		cost = time.Since(t)
	}()

	if samples == nil || mixer.Len() == 0 {
		goto __slient__
	}

	for n < len(p) {
		if r.bufPos >= r.bufSize {
			samples.BeZero()

			mixer.Stream(samples)
			r.resample.Stream(samples)

			if samples.LastNbSamples == 0 {
				// err = samples.LastErr
				goto __slient__
			}

			for i = 0; i < len(lines); i++ {
				if lines[i].IsDeleted() {
					RemoveLine(lines[i])
					i--
					continue
				}
				if playCallback != nil {
					playCallback(lines[i], samples)
				}
			}

			r.bufSize = samples.LastSamplesSize()
			r.bufPos = 0

			if slient {
				goto __slient__
			}
		}

		for i = 0; i < r.bufSize && n < len(p); i += 2 {
			p[n+0] = samples.RawData[0][r.bufPos]
			p[n+1] = samples.RawData[0][r.bufPos+1]
			p[n+2] = samples.RawData[1][r.bufPos]
			p[n+3] = samples.RawData[1][r.bufPos+1]
			n += 4
			r.bufPos += 2
		}
	}

	// err = samples.LastErr

	return

__slient__:
	zeroBuf(p)
	n = len(p)
	return

}

func zeroBuf(p []byte) {
	for i := 0; i < len(p); i++ {
		p[i] = 0
	}
}
