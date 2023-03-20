package detector

import (
	"net"
	"net/netip"
	"time"

	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/mutexer"
	"github.com/zwcway/castserver-go/pusher"
	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/castserver-go/web/websockets"
	"golang.org/x/net/ipv4"

	"go.uber.org/zap"
)

type recvData struct {
	pack *protocol.Package
	src  *net.UDPAddr
}

var (
	conn        *net.UDPConn
	log         *zap.Logger
	recvMessage chan *recvData
)

func ResponseServerInfo(sp *speaker.Speaker) {
	if sp == nil {
		return
	}
	sr := &ServerResponse{
		Ver:  1,
		Type: ST_Response,
		Addr: config.ServerListen.AddrPort.Addr(),
		Port: config.ServerListen.AddrPort.Port(),
	}
	p, err := sr.Pack()
	if err != nil {
		log.Error("send server info package invalid", zap.Error(err))
		return
	}

	n, err := conn.WriteToUDP(p.Bytes(), sp.UDPAddr())
	if err != nil {
		log.Error("send server info failed", zap.Error(err))
		return
	}
	if n != p.DataSize() {
		log.Error("send server info error", zap.Int("sended", n), zap.Int("size", p.DataSize()))
	}
}

func MulicastServerInfo(st ServerType) {
	sr := &ServerResponse{
		Ver:  1,
		Type: st,
		Addr: config.ServerListen.AddrPort.Addr(),
		Port: config.ServerListen.AddrPort.Port(),
	}
	p, err := sr.Pack()
	if err != nil {
		log.Error("send server info package invalid", zap.Error(err))
		return
	}

	addrPort := netip.AddrPortFrom(config.MulticastAddress, config.MulticastPort)
	n, err := conn.WriteToUDPAddrPort(p.Bytes(), addrPort)
	if err != nil {
		log.Error("send server info failed", zap.Error(err))
		return
	}
	if n != p.DataSize() {
		log.Error("send server info error", zap.Int("sended", n), zap.Int("size", p.DataSize()))
	}
}

func readUDP(p *recvData) {
	switch p.pack.Type() {
	default:
		return
	case protocol.PT_UNKNOWN:
		return
	case protocol.PT_ServerMutexRequest:
		if string(p.pack.Bytes()) == mutexer.TAG {
			conn.Write([]byte(mutexer.RSP))
		}
		return
	case protocol.PT_SpeakerInfo:
	}

	res := &SpeakerResponse{}
	err := res.Unpack(p.pack)
	if err != nil {
		log.Error("receive data is invalid", zap.Int("len", p.pack.Size()), zap.String("from", p.src.String()), zap.Error(err))
		return
	}
	if res.Addr.String() != p.src.IP.String() {
		log.Error("receive ip is wrong", zap.String("need", res.Addr.String()), zap.String("real", p.src.String()))
		return
	}

	if err = CheckSpeaker(res); err != nil {
		log.Error("invalid speaker", zap.String("from", p.src.String()))
	}

	return
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

		if d == nil || d.pack.Size() == 0 || d.src == nil {
			log.Info("channel closed")
			return
		}
		readUDP(d)
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
			log.Error("ReadFromUDP failed", zap.Error(err))
			return
		}

		if len(recvMessage) == cap(recvMessage) {
			log.Error("receive queue full")
			continue
		}

		recvMessage <- &recvData{protocol.FromBinary(buffer[:numBytes]), src}
	}
}

func listenUDP(ctx utils.Context) error {
	var err error
	addrPort := utils.UDPAddrFromAddr(&config.MulticastAddress, config.MulticastPort)
	conn, err = net.ListenMulticastUDP("udp", config.ServerListen.Iface, addrPort)
	if err != nil {
		return err
	}
	pc := ipv4.NewPacketConn(conn)

	if err := pc.SetMulticastLoopback(true); err != nil {
		log.Error("SetMulticastLoopback error:%v\n", zap.Error(err))
	}

	log.Info("start listen on " + addrPort.String())

	conn.SetReadBuffer(config.ReadBufferSize)

	MulicastServerInfo(ST_Start)

	go readUDPRoutine(config.ReadBufferSize)

	go readChanRoutine(ctx)

	return nil
}

func onlineCheckRoutine(ctx utils.Context) {
	ticker := time.NewTicker(time.Duration(config.ReadBufferSize) * time.Second)

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

			websockets.BroadcastSpeakerEvent(sp, websockets.Event_SP_Offline)
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
	MulicastServerInfo(ST_Exit)

	conn.Close()
}
