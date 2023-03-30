package pusher

import (
	config "github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	utils "github.com/zwcway/castserver-go/common/utils"
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
			log.Error("read from speaker failed", lg.Uint("speaker", uint64(sp.ID)), lg.Error(err))
			return
		}
		ip := addrPort.String()
		need := sp.UDPAddr().String()
		if ip != need {
			log.Error("received a invalid ip", lg.String("from", ip), lg.String("need", need))
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
