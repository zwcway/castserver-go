package decoder

import (
	"fmt"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

type Decoder struct {
	filePath string
	ctx      utils.Context
	log      *zap.Logger
	fs       *FileStream
	stream   beep.StreamSeekCloser
	format   beep.Format

	ctrl      *beep.Ctrl
	resampler *beep.Resampler
	volume    *effects.Volume

	isPlayed bool
}

func (d *Decoder) Close() error {
	if d == nil || d.fs == nil {
		return nil
	}
	d.filePath = ""
	d.isPlayed = false
	if d.stream != nil {
		speaker.Close()
		d.stream.Close()
	}
	return d.fs.Close()
}

func NewDecoder(ctx utils.Context) *Decoder {
	d := &Decoder{
		ctx: ctx,
		log: ctx.Logger("decoder"),
	}

	return d
}

func (d *Decoder) CurrentFile() string {
	return d.filePath
}

func (d *Decoder) Decode(path string) error {
	var err error

	d.filePath = path
	d.isPlayed = false

	d.fs = NewStream(d.ctx)
	err = d.fs.OpenFile(path)
	if err != nil {
		return err
	}
	buf := make([]byte, 512)
	_, err = d.fs.ReadAndRest(buf)
	if err != nil {
		return err
	}
	kind, err := filetype.Audio(buf)
	if err != nil {
		return err
	}
	if kind == types.Unknown {
		return fmt.Errorf("unknown file type for '%s'", path)
	}
	if kind == matchers.TypeFlac {
		d.stream, d.format, err = flac.Decode(d.fs)
	} else if kind == matchers.TypeMp3 {
		d.stream, d.format, err = mp3.Decode(d.fs)
	} else {
		return fmt.Errorf("unknown file type %s", kind.MIME.Value)
	}

	d.ctrl = &beep.Ctrl{Streamer: beep.Loop(-1, d.stream)}
	d.resampler = beep.ResampleRatio(4, 1, d.ctrl)
	d.volume = &effects.Volume{Streamer: d.resampler, Base: 2}

	return err
}

func (d *Decoder) controlBack(cb func()) {
	if d.isPlayed {
		speaker.Lock()
	}
	cb()
	if d.isPlayed {
		speaker.Unlock()
	}
}

func (d *Decoder) LocalPlay() {
	if d != nil && d.volume != nil && !d.isPlayed {
		speaker.Init(d.format.SampleRate, d.format.SampleRate.N(time.Second/30))
		speaker.Play(d.volume)
		d.isPlayed = true
		d.Unpause()
	}
}

func (d *Decoder) Pause() {
	if d != nil && d.ctrl != nil && d.isPlayed {
		d.controlBack(func() {
			d.ctrl.Paused = true
		})
	}
}

func (d *Decoder) Unpause() {
	if d != nil && d.ctrl != nil && d.isPlayed {
		d.controlBack(func() {
			d.ctrl.Paused = false
		})
	}
}
func (d *Decoder) IsPaused() bool {
	if d != nil && d.ctrl != nil && d.isPlayed {
		return d.ctrl.Paused
	}
	return false
}

func (d *Decoder) Seek(pos time.Duration) (err error) {
	if d == nil || d.stream == nil {
		return nil
	}
	d.controlBack(func() {
		newPos := d.format.SampleRate.N(pos)
		if newPos >= d.stream.Len() {
			newPos = d.stream.Len() - 1
		}
		err = d.stream.Seek(newPos)
	})

	return nil
}

func (d *Decoder) Channels() int {
	if d != nil {
		return d.format.NumChannels
	}
	return 0
}

func (d *Decoder) SampleRate() int {
	if d != nil {
		return int(d.format.SampleRate)
	}
	return 0
}

func (d *Decoder) Position() int {
	if d != nil && d.stream != nil {
		return d.stream.Position()
	}
	return 0
}
func (d *Decoder) Duration() time.Duration {
	if d != nil && d.stream != nil {
		return d.format.SampleRate.D(d.stream.Position())
	}
	return 0
}

func (d *Decoder) TotalDuration() time.Duration {
	if d != nil && d.stream != nil {
		return d.format.SampleRate.D(d.stream.Len())
	}
	return 0
}

func (d *Decoder) Volume() float64 {
	if d != nil && d.volume != nil {
		return d.volume.Volume
	}
	return 0
}

func (d *Decoder) Speed() float64 {
	if d != nil && d.resampler != nil {
		return d.resampler.Ratio()
	}
	return 0
}
func (d *Decoder) SetSpeed(ratio float64) {
	if d != nil && d.resampler != nil {
		d.controlBack(func() {
			d.resampler.SetRatio(d.resampler.Ratio() * ratio)
		})
	}
}
