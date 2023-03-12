package pusher

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"

	"go.uber.org/zap"
)

func closeChan(queue chan speaker.QueueData) {
	for _, sp := range speaker.AllSpeakers() {
		if sp.Queue == queue {
			Disconnect(sp)
		}
	}

	close(queue)
}

func pushRoutine(queue chan speaker.QueueData) {
	defer closeChan(queue)

	var d speaker.QueueData
	for {
		select {
		case <-context.Done():
			return
		case d = <-queue:
		}

		d.Speaker.Statistic.Queue -= uint32(len(d.Data))

		err := d.Speaker.WriteUDP(d.Data)
		if err != nil {
			log.Error("push to speaker error", zap.Error(err))
			continue
		}
	}
}

func PushToLineChannel(line *speaker.Line, ch audio.Channel, data []byte) {
	buf := ServerPush{
		Ver:     1,
		Seq:     1,
		Time:    1,
		Samples: data,
	}
	p, err := buf.Pack()
	if err != nil {
		return
	}

	data = p.Bytes()

	for _, sp := range line.SpeakersByChannel(ch) {
		Push(sp, data)
	}
}

func Push(sp *speaker.Speaker, data []byte) {
	queue := sp.Queue
	if queue == nil {
		// log.Error("speaker not connected", zap.String("speaker", sp.String()))
		return
	}
	if len(queue) == cap(queue) {
		log.Error("send queue full", zap.Uint32("speaker", uint32(sp.Id)), zap.Int("size", len(queue)))
		return
	}
	sp.Statistic.Queue += uint32(len(data))

	queue <- speaker.QueueData{Speaker: sp, Data: data}
}
