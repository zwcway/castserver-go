package localspeaker

import (
	"sync"
	"unsafe"

	"github.com/hajimehoshi/oto"
	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/decoder/pipeline"
)

var (
	mu      sync.Mutex
	pl      *pipeline.PipeLine
	buf     []byte
	context *oto.Context
	player  *oto.Player
	done    chan struct{}
)

func Init() error {
	mu.Lock()
	defer mu.Unlock()

	Close()

	format := &audio.Format{
		SampleRate: audio.AudioRate_44100,
		Layout:     audio.ChannelLayout20,
		SampleBits: audio.AudioBits_16LEF,
	}

	pipeline.SetOutputFormat(format)

	samples := pl.Buffer()
	size := 2 * samples.Size * format.SampleBits.Size()
	buf = make([]byte, size)

	var err error
	context, err = oto.NewContext(format.SampleRate.ToInt(), format.Layout.Count, format.SampleBits.Size(), size)

	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker")
	}
	player = context.NewPlayer()

	done = make(chan struct{})

	go func() {
		for {
			select {
			default:
				write()
			case <-done:
				return
			}
		}
	}()

	return nil
}

func Close() {
	if player != nil {
		if done != nil {
			done <- struct{}{}
			done = nil
		}
		player.Close()
		context.Close()
		pl.Close()
		player = nil
	}
}

func Lock() {
	mu.Lock()
}

func Unlock() {
	mu.Unlock()
}

func write() {
	mu.Lock()
	pl.Stream()
	mu.Unlock()

	pbuf := pl.Buffer()

	for c := 0; c < pbuf.Format.Layout.Count && c < 2; c++ {
		for i := 0; i < pbuf.Size; i++ {
			val := pbuf.Buffer[c][i]

			valInt16 := *(*int16)(unsafe.Pointer(&val))

			buf[i*4+c*2+0] = byte(valInt16)
			buf[i*4+c*2+1] = byte(valInt16 >> 8)
		}
	}

	player.Write(buf)
}
