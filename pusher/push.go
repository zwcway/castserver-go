package pusher

import (
	"sync"
	"time"

	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/speaker"

	"go.uber.org/zap"
)

type spQueueCtrl struct {
	exitC chan struct{}
	delay time.Duration
	sp    *speaker.Speaker
}

var spQueueCtrlList map[*speaker.Speaker]spQueueCtrl = make(map[*speaker.Speaker]spQueueCtrl)
var refreshLock sync.Mutex

func pushRoutine(queue spQueueCtrl) {
	var d speaker.QueueData
	for {
		select {
		case <-context.Done():
			Disconnect(queue.sp)
			close(queue.exitC)
			return
		case <-queue.exitC:
			close(queue.sp.Queue)
			close(queue.exitC)
			return
		case d = <-queue.sp.Queue:
		}
		if d.Speaker.IsDeleted() {
			continue
		}

		// TODO 按MTU拆包

		d.Speaker.Statistic.Queue -= uint32(len(d.Data))

		err := d.Speaker.WriteUDP(d.Data)
		if err != nil {
			log.Error("push to speaker error", zap.Error(err))
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
	refreshLock.Lock()
	defer refreshLock.Unlock()

	ctrl, ok := spQueueCtrlList[sp]
	if !ok {
		ctrl = spQueueCtrl{
			exitC: make(chan struct{}),
			sp:    sp,
		}
		spQueueCtrlList[sp] = ctrl
	}

	if sp.Queue != nil {
		ctrl.exitC <- struct{}{}
	}

	sp.Queue = make(chan speaker.QueueData, config.ReadQueueSize+bufSizeWithDelay(delay, sp.Format()))

	go pushRoutine(ctrl)
}
