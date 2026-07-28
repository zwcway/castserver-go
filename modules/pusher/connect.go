package pusher

import (
	"net"
	"time"

	"github.com/zwcway/castserver-go/common/config"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
)

// 创建数据连接
func Connect(sp *speaker.Speaker) error {
	if sp.Conn != nil {
		return nil
	}

	addr := config.ServerListen.AddrPort.Addr()
	port := config.ServerListen.AddrPort.Port()
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
	log.Info("connect speaker success", lg.Time("conn", sp.ConnTime))

	refreshPushQueue(sp, 0)

	go receiveSpeakerRoutine(sp)

	sp.Dispatch("speaker connected")

	return nil
}

func Disconnect(sp *speaker.Speaker) error {
	if sp == nil || sp.Conn == nil {
		return nil
	}
	close(sp.Queue)
	sp.Queue = nil
	// 关闭连接
	sp.Conn.Close()
	sp.Conn = nil

	sp.Dispatch("speaker disconnected")

	return nil
}

func connectionTest(sp *speaker.Speaker, udpAddr *net.UDPAddr) bool {
	if sp == nil {
		return false
	}
	conn, err := net.DialUDP("udp", udpAddr, sp.UDPAddr())
	if err != nil {
		log.Error("dial speaker error", lg.Error(err))
		return false
	}
	defer conn.Close()

	n, err := conn.Write([]byte{byte(protocol.PT_Ping)})
	if err != nil {
		log.Error("ping speaker error", lg.Error(err))
		return false
	}
	if n != 1 {
		log.Error("ping speaker error", lg.Int("writed", int64(n)), lg.Int("want", 1))
		return false
	}

	buf := make([]byte, 16)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err = conn.Read(buf)
	if err != nil {
		log.Info("read speaker error", lg.Error(err))
		return false
	}
	if buf[0] != byte(protocol.PT_Pong) {
		log.Info("read speaker pong error", lg.Int("size", int64(n)))
		return false
	}

	return true
}
