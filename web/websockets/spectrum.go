package websockets

import (
	"runtime"
	"sync"
	"time"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/go-jsonpack"
)

var ticker = time.NewTicker(50 * time.Millisecond)
var ctlSignal chan struct{}
var wg sync.WaitGroup
var locker sync.Mutex

var spectrumRunning = false

type eventService struct {
	arg int
	evt uint8
	se  stream.SpectrumElement
}

var services = []*eventService{}

type notifySpectrum struct {
	LevelMeter [2]float32 `jp:"l"`
	Spectrum   []float32  `jp:"s"`
}

func startSpectumRoutine() {
	locker.Lock()
	defer locker.Unlock()
	if ctlSignal != nil {
		return
	}
	ctlSignal = make(chan struct{}, 2)
	wg.Add(1)
	go spectrumRoutine()
}

func stopSpectumRoutine() {
	locker.Lock()
	defer locker.Unlock()
	if ctlSignal == nil {
		return
	}
	ctlSignal <- struct{}{}
	wg.Wait()
	close(ctlSignal)
	ctlSignal = nil
	spectrumRunning = false
}

func appendSpectrum(evt uint8, arg int) {
	locker.Lock()
	defer locker.Unlock()

	for _, es := range services {
		if es.evt == evt && es.arg == arg {
			return
		}
	}

	es := eventService{
		arg: arg,
		evt: evt,
	}

	switch evt {
	case Event_Line_Spectrum, Event_Line_LevelMeter:
		lineSpectrum(&es)
	case Event_SP_LevelMeter, Event_SP_Spectrum:
		speakerSpectrum(&es)
	default:
		return
	}
	services = append(services, &es)
}

func removeSpectrum(evt uint8, arg int) {
	locker.Lock()
	defer locker.Unlock()
	var (
		es *eventService
	)
	for i, e := range services {
		if e.evt == evt && arg == e.arg {
			es = e
			services = utils.SliceRemove(services, i)
			break
		}
	}

	if es == nil {
		return
	}
	if es.se != nil {
		es.se.Off()
	}

	switch evt {
	case Event_Line_Spectrum, Event_Line_LevelMeter:
		log.Info("stop line spectrum routine")
	case Event_SP_LevelMeter, Event_SP_Spectrum:
		log.Info("stop speaker spectrum routine")
	}

}

func lineSpectrum(es *eventService) {
	line := speaker.FindLineByID(speaker.LineID(es.arg))
	if line == nil {
		return
	}
	if es.se = line.Input.SpectrumEle; es.se == nil {
		return
	}
	if es.se.IsOn() {
		return
	}
	es.se.On()

	log.Info("start line spectrum")
}

func speakerSpectrum(es *eventService) {
	sp := speaker.FindSpeakerByID(speaker.SpeakerID(es.arg))
	if sp == nil || sp.SpectrumEle == nil {
		return
	}
	es.se = sp.SpectrumEle

	if es.se.IsOn() {
		return
	}
	es.se.On()

	log.Info("start speaker spectrum")

}

func spectrumRoutine() {
	defer wg.Done()

	locker.Lock()
	if spectrumRunning {
		locker.Unlock()
		return
	}
	spectrumRunning = true
	locker.Unlock()

	log.Info("start spectrum routine")

	for {
		select {
		case <-ctx.Done(): // 全局退出
			return
		case <-ctlSignal: // routine 退出
			log.Info("stop spectrum routine")
			return
		case <-ticker.C: // 定时器
		}

		runtime.Gosched()
		for _, a := range services {
			if a.se == nil {
				continue
			}
			st := a.se.Spectrum()

			// if len(st) == 0 && ls.LevelMeter() == 0 {
			// 	continue
			// }

			resp := notifySpectrum{
				LevelMeter: [2]float32{float32(a.arg), float32(a.se.LevelMeter())},
				Spectrum:   make([]float32, len(st)),
			}

			for i := 0; i < len(st); i++ {
				resp.Spectrum[i] = float32(st[i])
			}

			msg, err := jsonpack.Marshal(resp)
			if err == nil {
				Broadcast(a.evt, 0, a.arg, msg)
			}
		}
	}
}
