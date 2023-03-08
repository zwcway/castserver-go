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
	done    chan int
	quit    chan int
	size    int
)

func Init() error {
	format := &audio.Format{
		SampleRate: audio.AudioRate_44100,
		Layout:     audio.ChannelLayout20,
		SampleBits: audio.AudioBits_S16LE,
	}

	pl = pipeline.Default()
	if pl == nil {
		return errors.New("please init pipeline first")
	}
	pl.EleResample().SetFormat(format)

	mu.Lock()

	samples := pl.Buffer()
	size = format.Bytes()
	bufSize := samples.Size * size
	buf = make([]byte, bufSize)

	if player != nil {
		mu.Unlock()
		return nil
	}
	mu.Unlock()

	Close()

	var err error
	context, err = oto.NewContext(format.SampleRate.ToInt(), format.Layout.Count, format.SampleBits.Size(), bufSize)

	if err != nil {
		return errors.Wrap(err, "failed to initialize speaker")
	}
	player = context.NewPlayer()
	if player == nil {
		return errors.New("create player failed")
	}

	done = make(chan int, 2)
	quit = make(chan int, 2)
	go func() {
		for {

			select {
			default:
				write()
			case <-done:
				quit <- 1
				return
			}
		}
	}()

	return nil
}

func Close() {
	// mu.Lock()
	// defer mu.Unlock()

	if done != nil {
		done <- 1
		<-quit
		done = nil
		quit = nil
	}
	if player != nil {
		player.Close()
		context.Close()
		player = nil
	}
}

func write() {
	// mu.Lock()
	// defer mu.Unlock()

	pl.Stream()

	pbuf := pl.Buffer()

	if len(buf) < size*pbuf.Size {
		return
	}

	for c := 0; c < pbuf.Format.Layout.Count && c < 2; c++ {
		for i := 0; i < pbuf.Size && i*4+c*2 < len(buf); i++ {
			val := pbuf.Buffer[c][i]

			valInt16 := *(*int16)(unsafe.Pointer(&val))

			buf[i*4+c*2+0] = byte(valInt16)
			buf[i*4+c*2+1] = byte(valInt16 >> 8)
		}
	}

	player.Write(buf)
}
