package detector

import (
	"net"
	"time"

	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/modules/mutexer"
	"github.com/zwcway/castserver-go/modules/pusher"
	"github.com/zwcway/castserver-go/modules/web/websockets"
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

type recvData struct {
	message []byte
	src     *net.UDPAddr
}

var (
	conn        *net.UDPConn
	log         *zap.Logger
	recvMessage chan *recvData
)

func sendServerInfo(sp *speaker.Speaker) {

}

func readUDP(buf []byte, n int, src *net.UDPAddr) {
	if n == len(mutexer.TAG) && string(buf) == mutexer.TAG {
		conn.Write([]byte(mutexer.RSP))
		return
	}

	if n == packageSize(src.AddrPort().Addr().Is6()) {
		res, err := unPack(buf)
		if err != nil {
			log.Error("receive data is invalid", zap.Int("len", n), zap.String("from", src.String()))
			return
		}
		if res.Addr.String() != src.String() {
			log.Error("receive ip is wrong", zap.String("need", res.Addr.String()), zap.String("real", src.String()))
			return
		}

		if err = checkSpeaker(res); err != nil {
			log.Error("invalid speaker", zap.String("from", src.String()))
		}

		return
	}
}

func readChanRoutine(ctx utils.Context) {
	defer close(recvMessage)

	var d *recvData

	for {
		select {
		case d = <-recvMessage:
		case <-ctx.Done():
			return
		}

		if d == nil || len(d.message) == 0 || d.src == nil {
			log.Info("channel closed")
			return
		}
		readUDP(d.message, len(d.message), d.src)
	}
}

func readUDPRoutine(buferSize int) {
	buffer := make([]byte, buferSize)

	for {
		numBytes, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if utils.IsConnectCloseError(err) {
				log.Info("connection closed")
				return
			}
			log.Fatal("ReadFromUDP failed", zap.Error(err))
			return
		}

		if len(recvMessage) == cap(recvMessage) {
			log.Error("receive queue full")
			continue
		}

		recvMessage <- &recvData{buffer[:numBytes], src}
	}
}

func listenUDP(ctx utils.Context) error {
	var err error
	addrPort := utils.UDPAddrFromAddr(&config.MulticastAddress, config.MulticastPort)
	conn, err = net.ListenMulticastUDP("udp", config.DetectInterface, addrPort)
	if err != nil {
		return err
	}

	log.Info("start listen on " + addrPort.String())

	conn.SetReadBuffer(config.MaxReadBufferSize)

	go readUDPRoutine(config.MaxReadBufferSize)

	go readChanRoutine(ctx)

	return nil
}

func onlineCheckRoutine(ctx utils.Context) {
	ticker := time.NewTicker(time.Duration(config.MaxReadBufferSize) * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		speaker.All(func(sp *speaker.Speaker) {
			if !sp.IsOnline() {
				return
			}
			sp.Timeout--
			if sp.Timeout > 0 {
				return
			}

			log.Info("speaker is offline.", zap.String("speaker", sp.String()))
			sp.SetOffline()
			pusher.Disconnect(sp)

			sp.Timeout = 0

			websockets.BroadcastSpeakerEvent(sp, websockets.Event_SP_DETECTED)
		})
	}
}

type detectModule struct{}

var Module = detectModule{}

func (detectModule) Init(ctx utils.Context) error {
	log = ctx.Logger("detect")

	recvMessage = make(chan *recvData, 100)

	err := listenUDP(ctx)
	if err != nil {
		return err
	}

	go onlineCheckRoutine(ctx)

	return nil
}

func (detectModule) DeInit() {
	conn.Close()
}
