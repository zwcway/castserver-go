package pusher

import (
	"net"
	"time"

	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
	"go.uber.org/zap"
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
	log.Info("connect speaker success", zap.Time("conn", sp.ConnTime))

	refreshPushQueue(sp, 0)

	go receiveSpeakerRoutine(sp)

	bus.Dispatch("speaker connected", sp)

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

	bus.Dispatch("speaker disconnected", sp)

	return nil
}

func connectionTest(sp *speaker.Speaker, udpAddr *net.UDPAddr) bool {
	if sp == nil {
		return false
	}
	conn, err := net.DialUDP("udp", udpAddr, sp.UDPAddr())
	if err != nil {
		log.Error("dial speaker error", zap.Error(err))
		return false
	}
	defer conn.Close()

	n, err := conn.Write([]byte{byte(protocol.PT_Ping)})
	if err != nil {
		log.Error("ping speaker error", zap.Error(err))
		return false
	}
	if n != 1 {
		log.Error("ping speaker error", zap.Int("writed", n), zap.Int("want", 1))
		return false
	}

	buf := make([]byte, 16)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	n, err = conn.Read(buf)
	if err != nil {
		log.Info("read speaker error", zap.Error(err))
		return false
	}
	if buf[0] != byte(protocol.PT_Pong) {
		log.Info("read speaker pong error", zap.Int("size", n))
		return false
	}

	return true
}
