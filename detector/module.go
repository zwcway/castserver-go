package detector

import (
	"net"
	"net/netip"
	"time"

	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/protocol"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/mutexer"
	"github.com/zwcway/castserver-go/pusher"
	"golang.org/x/net/ipv4"
)

type recvData struct {
	pack *protocol.Package
	src  *net.UDPAddr
}

var (
	conn        *net.UDPConn
	ctx         utils.Context
	log         lg.Logger
	recvMessage chan *recvData
	wg          utils.WaitGroup
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
		log.Error("send server info package invalid", lg.Error(err))
		return
	}
	addr, err := netip.ParseAddr(sp.Ip)
	if err != nil {
		log.Error("speaker ip invalid", lg.String("speaker", sp.String()), lg.String("ip", sp.Ip), lg.Error(err))
		return
	}
	addrPort := netip.AddrPortFrom(addr, config.MulticastPort)

	n, err := conn.WriteToUDPAddrPort(p.Bytes(), addrPort)
	if err != nil {
		log.Error("send server info failed", lg.Error(err))
		return
	}
	if n != p.DataSize() {
		log.Error("send server info error", lg.Int("sended", int64(n)), lg.Int("size", int64(p.DataSize())))
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
		log.Error("send server info package invalid", lg.Error(err))
		return
	}

	addrPort := netip.AddrPortFrom(config.MulticastAddress, config.MulticastPort)
	n, err := conn.WriteToUDPAddrPort(p.Bytes(), addrPort)
	if err != nil {
		log.Error("send server info failed", lg.Error(err))
		return
	}
	if n != p.DataSize() {
		log.Error("send server info error", lg.Int("sended", int64(n)), lg.Int("size", int64(p.DataSize())))
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
		log.Error("receive data is invalid", lg.Int("len", int64(p.pack.Size())), lg.String("from", p.src.String()), lg.Error(err))
		return
	}
	if res.Addr.String() != p.src.IP.String() {
		log.Error("receive ip is wrong", lg.String("need", res.Addr.String()), lg.String("real", p.src.String()))
		return
	}

	if err = CheckSpeaker(res); err != nil {
		log.Error("invalid speaker", lg.String("from", p.src.String()))
	}

}

func readChanRoutine(ctx utils.Context, done <-chan struct{}) {
	var d *recvData

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case d = <-recvMessage:
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
			log.Error("ReadFromUDP failed", lg.Error(err))
			return
		}

		if len(recvMessage) == cap(recvMessage) {
			log.Error("receive queue full")
			continue
		}

		log.Debug("receive data", lg.String("from", src.String()), lg.Binary("data", buffer[:numBytes]))
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
		log.Error("SetMulticastLoopback error:%v\n", lg.Error(err))
	}

	log.Info("start listen on " + config.ServerListen.AddrPort.String())

	conn.SetReadBuffer(config.ReadBufferSize)

	MulicastServerInfo(ST_Start)

	wg.Go(func(<-chan struct{}) {
		readUDPRoutine(config.ReadBufferSize)
	})

	wg.Go(func(done <-chan struct{}) {
		readChanRoutine(ctx, done)
	})

	return nil
}

func onlineCheckRoutine(ctx utils.Context, done <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(config.ReadBufferSize) * time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
		}

		speaker.All(func(sp *speaker.Speaker) {
			if !sp.IsOnline() {
				return
			}
			sp.Timeout--
			if sp.Timeout > 0 {
				return
			}

			log.Info("speaker is offline.", lg.String("speaker", sp.String()))
			sp.SetOffline()
			pusher.Disconnect(sp)

			sp.Timeout = 0

			bus.Dispatch("speaker offline", sp)
		})
	}
}

type detectModule struct{}

var Module = detectModule{}

func (detectModule) Init(uctx utils.Context) error {
	ctx = uctx
	log = ctx.Logger("detect")

	recvMessage = make(chan *recvData, 100)

	return nil
}

func (detectModule) Start() error {

	err := listenUDP(ctx)
	if err != nil {
		return err
	}

	wg.Go(func(done <-chan struct{}) {
		onlineCheckRoutine(ctx, done)
	})

	return nil
}

func (detectModule) DeInit() {
	MulicastServerInfo(ST_Exit)

	conn.Close()
	wg.Wait()
	close(recvMessage)
}
