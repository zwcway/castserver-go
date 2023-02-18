package pusher

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

func closeChan(queue *chan sendQueue) {
	for sp, q := range queueSpeaker {
		if q == queue {
			Disconnect(sp)
		}
	}

	close(*queue)
}

func pushRoutine(queue *chan sendQueue) {
	defer closeChan(queue)

	for {
		var d sendQueue
		select {
		case <-context.Done():
			return
		case d = <-*queue:
		}

		numBytes, err := d.sp.Conn.Write(d.data)
		if err != nil {
			log.Error("push to speaker error", zap.Error(err))
			continue
		}
		if len(d.data) != numBytes {
			log.Warn("push uncomplete", zap.Int("pushed", numBytes), zap.Int("data", len(d.data)))
		}
	}
}

func receiveData(sp *speaker.Speaker) {
	receiveBuffer := make([]byte, config.MaxReadBufferSize)

	numBytes, addrPort, err := sp.Conn.ReadFromUDPAddrPort(receiveBuffer)
	if err != nil {
		if utils.IsConnectCloseError(err) {
			return
		}
		log.Fatal("ReadFromUDP failed", zap.Error(err))
		return
	}
	ip := addrPort.String()
	need := sp.UDPAddr().String()
	if ip != need {
		log.Error("received a invalid ip", zap.String("from", ip), zap.String("need", need))
		return
	}
	_ = receiveBuffer[:numBytes]
}

func PushToChannel(ch audio.AudioChannel, data []byte) {
	speaker.All(func(s *speaker.Speaker) {
		if s.Channel == ch {
			Push(s, data)
		}
	})
}

func PushToLineChannel(l speaker.SpeakerLineID, ch audio.AudioChannel, data []byte) {
	sps, ok := speaker.SpeakersByLine(l)
	if !ok {
		return
	}

	for _, sp := range sps {
		if sp.Channel == ch {
			Push(sp, data)
		}
	}
}

func Push(sp *speaker.Speaker, data []byte) {
	queue, ok := queueSpeaker[sp]
	if !ok {
		// log.Fatal("speaker not connected", zap.String("speaker", sp.String()))
		return
	}
	*queue <- sendQueue{sp, data}
}
