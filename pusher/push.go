package pusher

import (
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
		if d.Speaker.IsDeleted() {
			continue
		}

		d.Speaker.Statistic.Queue -= uint32(len(d.Data))

		err := d.Speaker.WriteUDP(d.Data)
		if err != nil {
			log.Error("push to speaker error", zap.Error(err))
			continue
		}
	}
}
