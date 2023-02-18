package ssdp

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	config "github.com/zwcway/castserver-go/config"
	utils "github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
	"golang.org/x/net/ipv4"
)

const (
	MulticastAddrPort string = "239.255.255.250:1900"
)

type SSDPServer interface {
	ListenAndServe() error
	Close()
}
type ssdpServer struct {
	ctx        utils.Context
	log        *zap.Logger
	signal     chan int
	conn       *net.UDPConn
	listenAddr *net.UDPAddr
	iface      *net.Interface
	addrs      []*net.IPNet

	UUID   string
	Server string

	Devices  []string
	Services []string

	Location func(ip net.IP) string
}

func NewSSDPServer(ctx utils.Context, iface *net.Interface) (*ssdpServer, error) {
	if iface == nil {
		return nil, fmt.Errorf("interface nil")
	}

	addrs := utils.InterfaceAddrs(iface, nil)
	if addrs == nil {
		return nil, fmt.Errorf("there are no address on interface %s", iface.Name)
	}

	return &ssdpServer{
		ctx:    ctx,
		log:    ctx.Logger("dlna ssdp"),
		signal: make(chan int, 1),
		iface:  iface,
		addrs:  addrs,
		Location: func(ip net.IP) string {
			return ip.String()
		},
	}, nil
}

func (s *ssdpServer) Close() {
	s.signal <- 1 // 一切阻塞终将退出

	s.sendByeBye()
	close(s.signal)
	s.conn.Close()
}
func checkSSDPConfig(s *ssdpServer) bool {
	if s.Location == nil {
		return false
	}
	return true
}

func (s *ssdpServer) ListenAndServe() error {
	var err error
	s.listenAddr, err = net.ResolveUDPAddr("udp4", MulticastAddrPort)
	if err != nil {
		panic(err)
	}
	if !checkSSDPConfig(s) {
		panic("ssdp config invalid")
	}
	s.conn, err = net.ListenMulticastUDP("udp", config.DetectInterface, s.listenAddr)
	if err != nil {
		return err
	}
	pack := ipv4.NewPacketConn(s.conn)
	err = pack.SetMulticastTTL(2)
	if err != nil {
		return err
	}

	go s.readUdpRoutine()

	s.multicast()
	s.Close()
	return nil
}

func (s *ssdpServer) multicast() {
	tick := time.NewTicker(time.Duration(config.DLNANotifyInterval) * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.signal:
			return
		case <-tick.C:
		}
		for _, addr := range s.addrs {
			extHeads := map[string]string{
				"CACHE-CONTROL": fmt.Sprintf("max-age=%d", 5*config.DLNANotifyInterval/2),
				"LOCATION":      s.Location(addr.IP),
			}
			s.sendAlive(extHeads)
		}
	}
}
func (s *ssdpServer) readUdpRoutine() {
	bs := int(math.Max(65535, float64(config.MTU())))
	if bs <= 0 {
		bs = 65535
	}

	buf := make([]byte, bs)
	for {
		num, src, err := s.conn.ReadFromUDP(buf)

		if err != nil {
			if !utils.IsConnectCloseError(err) {
				s.log.Panic("abnormally exit", zap.Error(err))
			}
			// 全部退出
			s.signal <- 1
			return
		}
		go s.readRequestRoutine(buf[:num], src)
	}
}
func (s *ssdpServer) checkRequest(req *http.Request) bool {
	man := strings.Trim(req.Header.Get("man"), "\"")
	if req.Method != "M-SEARCH" || man != "ssdp:discover" {
		// s.log.Warn("invalid request", zap.String("method", req.Method), zap.String("man", man))
		return false
	}

	return true
}
func (s *ssdpServer) readMX(req *http.Request) int64 {
	if req.Header.Get("Host") == MulticastAddrPort {
		mxhd := req.Header.Get("mx")
		i, err := strconv.ParseUint(mxhd, 0, 8)
		if err == nil && i > 0 {
			return int64(i)
		}
		s.log.Warn("invalid mx header", zap.String("mx", mxhd))
	}
	return 1
}

func (s *ssdpServer) readSTS(req *http.Request) []string {
	st := req.Header.Get("ST")
	if st == "ssdp:all" {
		return s.ntList()
	}
	for _, nt := range s.ntList() {
		if st == nt {
			return []string{st}
		}
	}

	return nil
}

func (s *ssdpServer) ipnetContains(src net.IP) net.IP {
	for _, in := range s.addrs {
		if in.Contains(src) {
			return in.IP
		}
	}
	return nil
}
func (s *ssdpServer) readRequestRoutine(buf []byte, src *net.UDPAddr) {
	io := bufio.NewReader(bytes.NewReader(buf))
	req, err := http.ReadRequest(io)
	if err != nil {
		s.log.Warn("bad request from client", zap.String("from", src.String()), zap.Error(err))
		return
	}
	if !s.checkRequest(req) {
		return
	}
	mx := s.readMX(req)
	sts := s.readSTS(req)

	// s.log.Info("receive", zap.String("from", src.String()), zap.String("body", string(buf)))

	ip := s.ipnetContains(src.IP)
	if ip == nil {
		return
	}
	// 单播响应
	for _, st := range sts {
		resp := s.makeResponse(ip, st)
		s.send(resp, src, time.Duration(rand.Int63n(mx)))
	}
}

func (s *ssdpServer) makeUSN(nt string) string {
	if s.UUID == nt {
		return nt
	}
	return s.UUID + "::" + nt
}

func (s *ssdpServer) ntList() []string {
	list := make([]string, 2)
	list[0] = "upnp:rootdevice"
	list[1] = s.UUID

	list = append(list, s.Devices...)
	list = append(list, s.Services...)

	return list
}

func (s *ssdpServer) send(buf []byte, ip *net.UDPAddr, delay time.Duration) {
	// s.log.Info("send", zap.String("body", string(buf)))

	go func() {
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-s.signal:
				return
			case <-s.ctx.Done():
				s.Close()
				return
			}
		}
		num, err := s.conn.WriteToUDP(buf, ip)
		if err != nil {
			s.log.Fatal("write udp error", zap.Error(err))
		} else if num != len(buf) {
			s.log.Fatal("write udp size error", zap.Int("send", num), zap.Int("datalen", len(buf)))
		}
	}()
}

func appendHeaders(buf *bytes.Buffer, hd any) {
	switch hd.(type) {
	case map[string]string:
		for k, v := range hd.(map[string]string) {
			fmt.Fprintf(buf, "%s: %s\r\n", k, v)
		}
	case string:
		fmt.Fprint(buf, hd.(string))
	}
}

func (s *ssdpServer) makeResponse(ip net.IP, st string) []byte {
	head := map[string]string{
		"CACHE-CONTROL": fmt.Sprintf("max-age=%d", 5*config.DLNANotifyInterval/2),
		"EXT":           "",
		"LOCATION":      s.Location(ip),
		"SERVER":        s.Server,
		"ST":            st,
		"USN":           s.makeUSN(st),
	}
	buf := &bytes.Buffer{}
	appendHeaders(buf, "HTTP/1.1 200 OK\r\n")
	appendHeaders(buf, head)
	return buf.Bytes()
}

func (s *ssdpServer) makeNotify(nt, nts string, extHeads map[string]string) []byte {
	head := map[string]string{
		"HOST":   MulticastAddrPort,
		"NT":     nt,
		"NTS":    nts,
		"SERVER": s.Server,
		"USN":    s.makeUSN(nt),
	}

	buf := &bytes.Buffer{}
	appendHeaders(buf, "NOTIFY * HTTP/1.1\r\n")
	appendHeaders(buf, head)
	appendHeaders(buf, extHeads)
	appendHeaders(buf, "\r\n")
	return buf.Bytes()
}

func (s *ssdpServer) sendByeBye() {
	for _, nt := range s.ntList() {
		buf := s.makeNotify(nt, "ssdp:byebye", nil)
		s.send(buf, s.listenAddr, 0)
	}
}

func (s *ssdpServer) sendAlive(extHeads map[string]string) {
	for _, nt := range s.ntList() {
		buf := s.makeNotify(nt, "ssdp:alive", extHeads)
		s.send(buf, s.listenAddr, time.Duration(rand.Int63n(100*int64(time.Millisecond))))
	}
}
