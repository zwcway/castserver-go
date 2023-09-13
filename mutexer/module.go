package mutexer

import (
	"fmt"
	"net"
	"strings"
	"time"

	config "github.com/zwcway/castserver-go/common/config"
	log1 "github.com/zwcway/castserver-go/common/log"
	utils "github.com/zwcway/castserver-go/common/utils"
)

var (
	conn   *net.UDPConn
	log    log1.Logger
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
		log.Error("there are another server running. exiting.", log1.String("addr", src.String()))
		return fmt.Errorf("exiting")
	}

	return nil
}

func (mutexModule) Init(ctx utils.Context) error {
	log = ctx.Logger("mutexer")

	return nil
}

func (mutexModule) Start() error {
	return listenUDP()
}

func (mutexModule) DeInit() {
}
