package pusher

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"

	"go.uber.org/zap"
)

func closeChan(queue *chan QueueData) {
	for sp, q := range queueSpeaker {
		if q == queue {
			Disconnect(sp)
		}
	}

	close(*queue)
}

func pushRoutine(queue *chan QueueData) {
	defer closeChan(queue)

	for {
		var d QueueData
		select {
		case <-context.Done():
			return
		case d = <-*queue:
		}

		d.sp.Statistic.Queue -= uint32(len(d.data))

		err := d.sp.WriteUDP(d.data)
		if err != nil {
			log.Error("push to speaker error", zap.Error(err))
			continue
		}
	}
}

func PushToLineChannel(l speaker.LineID, ch audio.Channel, data []byte) {
	for _, sp := range speaker.FindSpeakersByChannel(l, ch) {
		Push(sp, data)
	}
}

func Push(sp *speaker.Speaker, data []byte) {
	queue, ok := queueSpeaker[sp]
	if !ok {
		// log.Error("speaker not connected", zap.String("speaker", sp.String()))
		return
	}
	if len(*queue) == cap(*queue) {
		log.Error("send queue full", zap.Uint32("speaker", uint32(sp.ID)), zap.Int("size", len(*queue)))
		return
	}
	sp.Statistic.Queue += uint32(len(data))

	*queue <- QueueData{sp, data}
}
