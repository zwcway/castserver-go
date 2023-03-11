package localspeaker

import (
	"sync"

	oto "github.com/hajimehoshi/oto/v2"
	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/decoder/element"
	"github.com/zwcway/castserver-go/utils"
)

var (
	mu           sync.Mutex
	context      *oto.Context
	player       oto.Player
	sampleReader *reader
	mixer        stream.MixerElement

	lines []*speaker.Line
)

func Init() error {
	mu.Lock()
	defer mu.Unlock()
	if player != nil {
		return nil
	}

	format := &audio.Format{
		SampleRate: audio.AudioRate_44100,
		Layout:     audio.ChannelLayout20,
		SampleBits: audio.AudioBits_S16LE,
	}

	var err error
	samples := 17640
	bufSize := samples * format.Bytes()
	mixer = element.NewMixer(nil)
	sampleReader = &reader{buf: stream.NewSamples(samples, format)}

	var readyChan chan struct{}

	context, readyChan, err = oto.NewContext(format.SampleRate.ToInt(), format.Layout.Count, format.SampleBits.Size())
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
	mixer.Add(line.Input.PipeLine)
	lines = append(lines, line)
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

func Play() {
	if player == nil || mixer == nil || mixer.Len() == 0 {
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
		mixer.Clear()
		lines = lines[:0]
	}

	return nil
}

type reader struct {
	buf   *stream.Samples
	total int
	pos   int
}

func (r *reader) Read(p []byte) (n int, err error) {
	rs := len(p) / 2 / 2

	pbuf := r.buf
	if pbuf == nil || pbuf.Format.Layout.Count < 1 {
		for i := 0; i < len(p); i++ {
			p[i] = 0
		}
		r.pos = 0
		r.total = 0
		return len(p), nil
	}

	r.buf.BeZero()

	var valInt16 int16
	n = 0
	for i := 0; i < rs; i++ {
		if r.pos >= r.total {
			mixer.Stream(pbuf)
			r.total = pbuf.LastSize
			r.pos = 0
			if pbuf.LastErr != nil || pbuf.LastSize == 0 {
				err = pbuf.LastErr
				return
			}
		}
		for ch := 0; ch < 2; ch++ {
			if pbuf.Format.Layout.Count > 1 {
				val := pbuf.Buffer[ch][r.pos]
				if val < -1 {
					val = -1
				}
				if val > +1 {
					val = +1
				}
				valInt16 = int16(val * (1<<15 - 1))
			}

			p[i*4+ch*2+0] = byte(valInt16)
			p[i*4+ch*2+1] = byte(valInt16 >> 8)
			n += 2
		}
		r.pos++
	}

	return
}
