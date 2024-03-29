package pusher

import (
	"time"

	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
)

type spQueueCtrl struct {
	exitC chan struct{}
	delay time.Duration
	sp    *speaker.Speaker
}

var spQueueCtrlList map[*speaker.Speaker]spQueueCtrl = make(map[*speaker.Speaker]spQueueCtrl)
var wg = utils.WaitGroup{}

func pushRoutine(queue spQueueCtrl, done <-chan struct{}) {
	var d speaker.QueueData
	for {
		select {
		case <-context.Done():
			Disconnect(queue.sp)
			close(queue.exitC)
			return
		case <-done:
			return
		case d = <-queue.sp.Queue:
		}
		if d.Speaker == nil || d.Speaker.IsDeleted() {
			continue
		}

		// TODO 按MTU拆包

		d.Speaker.Statistic.Queue -= uint32(len(d.Data))

		err := d.Speaker.WriteUDP(d.Data)
		if err != nil {
			// log.Error("push to speaker error", lg.Error(err))
			continue
		}
	}
}

func delayChanged(sp *speaker.Speaker, delay time.Duration) bool {
	ctrl, ok := spQueueCtrlList[sp]
	if !ok {
		return false
	}

	return ctrl.delay != delay
}

func refreshPushQueue(sp *speaker.Speaker, delay time.Duration) {
	ctrl, ok := spQueueCtrlList[sp]
	if !ok {
		ctrl = spQueueCtrl{
			exitC: make(chan struct{}),
			sp:    sp,
		}
		spQueueCtrlList[sp] = ctrl
	}

	wg.ExitAndWait()

	if sp.Queue != nil {
		close(sp.Queue)
	}

	sp.Queue = make(chan speaker.QueueData, config.ReadQueueSize+bufSizeWithDelay(delay, sp.Format()))

	wg.Go(func(done <-chan struct{}) {
		pushRoutine(ctrl, done)
	})
}
