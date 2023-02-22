package pusher

import (
	"net"
	"time"

	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"
)

type QueueData struct {
	sp   *speaker.Speaker
	data []byte
}

var queueList []chan QueueData
var queueIndex int = 0
var queueSpeaker map[*speaker.Speaker]*chan QueueData

func Connect(sp *speaker.Speaker) error {
	if sp.Conn != nil {
		return nil
	}

	addr := config.ServerAddrPort.Addr()
	port := config.ServerAddrPort.Port()
	udpAddr := utils.UDPAddrFromAddr(&addr, port)
	if addr.IsUnspecified() {
		udpAddr = nil
	}
	var err error
	sp.Conn, err = net.DialUDP("udp", udpAddr, sp.UDPAddr())
	if err != nil {
		return err
	}
	sp.ConnTime = time.Now()

	// 将设备添加到发送队列中
	var queue *chan QueueData
	if len(queueList) == queueIndex {
		queueIndex = 0
		queue = &queueList[queueIndex]
	} else {
		queue = &queueList[queueIndex]
		queueIndex++
	}
	queueSpeaker[sp] = queue

	go receiveSpeakerRoutine(sp)

	return nil
}

func Disconnect(sp *speaker.Speaker) error {
	if sp == nil || sp.Conn == nil {
		return nil
	}

	delete(queueSpeaker, sp)
	// 关闭连接
	sp.Conn.Close()
	sp.Conn = nil

	return nil
}
