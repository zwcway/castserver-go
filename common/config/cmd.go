package config

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	utils "github.com/zwcway/castserver-go/common/utils"

	"github.com/jedib0t/go-pretty/table"
	gap "github.com/muesli/go-app-paths"
	"go.uber.org/zap"
)

func readConfigFile(log *zap.Logger, opts map[string]string) error {
	var path string

	if v, ok := opts["c"]; ok {
		path, err := filepath.Abs(v)
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return err
		}

		return FromFile(log, path)
	}

	scope := gap.NewScope(gap.User, APPNAME)
	dirs, err := scope.ConfigDirs()
	if err != nil {
		return err
	}
	cd, err := os.Getwd()
	if err != nil {
		return err
	}
	dirs = append([]string{cd}, dirs...)

	for _, d := range dirs {
		path = filepath.Join(d, ConfigFile)
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return FromFile(log, path)
		}
	}
	return FromContent(log, []byte{})
}

func printInterfaces() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Index", "Name", "Addresses"})

	ifis := utils.Interfaces()
	for _, ifi := range ifis {
		addrs, _ := ifi.Addrs()
		addrStr := []string{}
		for _, addr := range addrs {
			if ip, _ := addr.(*net.IPNet); utils.CheckIPNet(ip) {
				addrStr = append(addrStr, ip.IP.String())
			}
		}
		t.AppendRow(table.Row{ifi.Index, ifi.Name, strings.Join(addrStr, ",")})
	}
	t.Render()
}

func FromOptions(log *zap.Logger, opts map[string]string) error {
	var (
		v  string
		ok bool
	)

	if err := readConfigFile(log, opts); err != nil {
		return err
	}

	// 覆盖配置文件

	if v, ok = opts["multicast-ip"]; ok {
		addr, err := netip.ParseAddr(v)
		if err != nil {
			return err
		}
		if !addr.IsMulticast() {
			return fmt.Errorf("%s is not a multicast ip", v)
		}
		MulticastAddress = addr
	}

	if v, ok = opts["multicast-port"]; ok {
		port, err := strconv.ParseInt(v, 0, 16)
		if err != nil {
			return err
		}
		if !utils.PortIsValid(uint16(port)) {
			return fmt.Errorf("port %s is invalid", v)
		}
		MulticastPort = uint16(port)
	}

	if v, ok = opts["i"]; ok {
		iface, addr := cmdInterface(v)

		if addr != nil {
			iface.AddrPort = netip.AddrPortFrom(*addr, ServerListen.AddrPort.Port())
			ServerListen = *iface
			iface.AddrPort = netip.AddrPortFrom(*addr, ReceiveListen.AddrPort.Port())
			ReceiveListen = *iface
			iface.AddrPort = netip.AddrPortFrom(*addr, HTTPListen.AddrPort.Port())
			HTTPListen = *iface
			iface.AddrPort = netip.AddrPortFrom(*addr, DLNAListen.AddrPort.Port())
			DLNAListen = *iface
		} else {
			ServerListen = *iface
			ReceiveListen = *iface
			HTTPListen = *iface
			DLNAListen = *iface
		}
	}

	if v, ok = opts["detect-interface"]; ok {
		iface, addr := cmdInterface(v)
		if iface == nil {
			return fmt.Errorf("")
		}

		if addr != nil {
			iface.AddrPort = netip.AddrPortFrom(*addr, ServerListen.AddrPort.Port())
		}
		ServerListen = *iface
	}

	return nil
}

func cmdInterface(v string) (iface *Interface, addr *netip.Addr) {
	iface = &Interface{}

	if utils.IsUint(v) {
		i, _ := strconv.ParseInt(v, 0, 32)
		iface.Iface = utils.InterfaceByIndex(int(i))
	} else {
		iface.Iface = utils.InterfaceByName(v)
	}

	if iface.Iface == nil {
		fmt.Printf("interface '%s' is invalid\n", v)
		printInterfaces()
		return nil, nil
	}

	ip := utils.InterfaceAddr(iface.Iface, false)
	addr = utils.IpNetToAddr(ip)

	return
}
