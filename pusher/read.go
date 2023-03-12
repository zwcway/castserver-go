package pusher

import (
	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

var receiveQueue chan speaker.QueueData

func receiveSpeakerRoutine(sp *speaker.Speaker) {

	for {
		receiveBuffer := make([]byte, config.ReadBufferSize)

		numBytes, addrPort, err := sp.Conn.ReadFromUDPAddrPort(receiveBuffer)
		if err != nil {
			if utils.IsConnectCloseError(err) {
				return
			}
			log.Error("read from speaker failed", zap.Uint32("speaker", uint32(sp.Id)), zap.Error(err))
			return
		}
		ip := addrPort.String()
		need := sp.UDPAddr().String()
		if ip != need {
			log.Error("received a invalid ip", zap.String("from", ip), zap.String("need", need))
			return
		}

		if len(receiveQueue) < cap(receiveQueue) {
			receiveQueue <- speaker.QueueData{
				Speaker: sp,
				Data:    receiveBuffer[:numBytes],
			}
		}
	}
}
