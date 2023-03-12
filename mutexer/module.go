package mutexer

import (
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"

	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

var (
	conn   *net.UDPConn
	log    *zap.Logger
	Module = mutexModule{}
)

type mutexModule struct{}

func listenUDP() error {
	var err error
	addrPort := utils.UDPAddrFromAddr(&config.MulticastAddress, config.MulticastPort)

	conn, err = net.ListenMulticastUDP("udp", config.ServerListen.Iface, addrPort)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	conn.SetReadBuffer(512)

	log.Info("sending mutex")

	numBytes, err := conn.WriteToUDP([]byte(TAG), addrPort)
	if err != nil {
		return err
	}
	if numBytes == 0 {
		return fmt.Errorf("send error")
	}

	buffer := make([]byte, 512)
	numBytes, src, err := conn.ReadFromUDP(buffer)
	if err != nil && !strings.HasSuffix(err.Error(), "i/o timeout") {
		return err
	}

	str := string(buffer[:numBytes])
	if str == RSP || str == TAG {
		log.Error("there are another server running. exiting.", zap.String("addr", src.String()))
		return fmt.Errorf("exiting")
	}

	return nil
}

func (mutexModule) Init(ctx utils.Context) error {
	log = ctx.Logger("mutexer")

	signal := ctx.Signal()

	err := listenUDP()

	if err != nil {
		signal <- syscall.SIGTERM
		return err
	}

	return nil
}

func (mutexModule) DeInit() {
}
