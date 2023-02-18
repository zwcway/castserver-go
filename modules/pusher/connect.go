package pusher

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"
)

type sendQueue struct {
	sp   *speaker.Speaker
	data []byte
}

var queueList []chan sendQueue
var queueIndex int = 0
var queueSpeaker map[*speaker.Speaker]*chan sendQueue

func Connect(sp *speaker.Speaker) error {
	if sp.Conn != nil {
		return nil
	}
	lAddr := utils.InterfaceAddr(config.DetectInterface, config.DetectUseIPV6)
	if lAddr == nil {
		return fmt.Errorf("get ip from interface error")
	}
	addr, ok := netip.AddrFromSlice(lAddr.IP)
	if !ok {
		return fmt.Errorf("invalid ip from interface %s", lAddr.String())
	}
	var err error
	sp.Conn, err = net.DialUDP("udp", utils.UDPAddrFromAddr(&addr, 0), sp.UDPAddr())
	if err != nil {
		return err
	}

	var queue *chan sendQueue
	if len(queueList) == queueIndex {
		queueIndex = 0
		queue = &queueList[queueIndex]
	} else {
		queue = &queueList[queueIndex]
		go pushRoutine(queue)
		queueIndex++
	}
	queueSpeaker[sp] = queue

	return nil
}

func Disconnect(sp *speaker.Speaker) error {
	if sp.Conn == nil {
		return nil
	}

	delete(queueSpeaker, sp)
	// 关闭连接
	sp.Conn.Close()
	sp.Conn = nil

	return nil
}
