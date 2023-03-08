package pusher

import (
	"net"
	"time"

	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
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

	if !connectionTest(sp, udpAddr) {
		return nil
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

func connectionTest(sp *speaker.Speaker, udpAddr *net.UDPAddr) bool {
	if sp == nil {
		return false
	}
	conn, err := net.DialUDP("udp", udpAddr, sp.UDPAddr())
	if err != nil {
		return false
	}
	defer conn.Close()

	req := ServerPush{}
	p, err := req.Pack()
	if err != nil {
		log.Error("pack error", zap.Error(err))
		return false
	}

	n, err := conn.Write(p.Bytes())
	if err != nil {
		log.Error("write speaker error", zap.Error(err))
		return false
	}
	if n != p.DataSize() {
		log.Error("write speaker error", zap.Int("writed", n), zap.Int("want", p.DataSize()))
		return false
	}

	buf := make([]byte, 16)
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	n, err = conn.Read(buf)
	if err != nil {
		log.Info("read speaker error", zap.Error(err))
		return false
	}

	return true
}
