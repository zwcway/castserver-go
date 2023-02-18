package utils

import (
	"encoding/binary"
	"net"
	"net/netip"
	"strings"
)

func IsConnectCloseError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "use of closed network connection")
}

func MacIsValid(m net.HardwareAddr) bool {
	return binary.BigEndian.Uint64(m) != 0
}

func PortIsValid(p uint16) bool {
	return int32(p) > 0 && int32(p) < 0xFFFF
}

func CheckInterface(ifi *net.Interface) bool {
	return ifi != nil && ifi.Flags&net.FlagLoopback == 0 && ifi.Flags&net.FlagMulticast != 0
}

func CheckIPNet(a *net.IPNet) bool {
	if a == nil {
		return false
	}
	ip := a.IP

	return !ip.IsLoopback() && !ip.IsMulticast() && !ip.IsInterfaceLocalMulticast() && !ip.IsLinkLocalMulticast() && !ip.IsLinkLocalUnicast()
}

func UDPAddrFromAddr(ip *netip.Addr, port uint16) *net.UDPAddr {
	if ip == nil {
		return nil
	}
	return &net.UDPAddr{
		IP:   ip.AsSlice(),
		Zone: ip.Zone(),
		Port: int(port),
	}
}

func InterfaceAddrs(iface *net.Interface, filter func(net.Addr) bool) []*net.IPNet {
	if iface == nil {
		return nil
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil
	}
	addrStr := []*net.IPNet{}
	for _, addr := range addrs {
		if filter != nil && !filter(addr) {
			continue
		}
		if ip, ok := addr.(*net.IPNet); ok && CheckIPNet(ip) {
			addrStr = append(addrStr, ip)
		} else if ip, ok := addr.(*net.IPAddr); ok {
			if len(ip.IP) == net.IPv4len {
				addrStr = append(addrStr, &net.IPNet{
					IP:   ip.IP,
					Mask: net.IPv4Mask(0, 0, 0, 0),
				})
			} else if len(ip.IP) == net.IPv6len {
				addrStr = append(addrStr, &net.IPNet{
					IP:   ip.IP,
					Mask: net.CIDRMask(0, 8*net.IPv6len),
				})
			}
		}
	}
	if len(addrStr) == 0 {
		return nil
	}

	return addrStr
}


func InterfaceAddr(iface *net.Interface, ipv6 bool) *net.IPNet {
	if iface == nil {
		return nil
	}
	addrs := InterfaceAddrs(iface, nil)
	for _, a := range addrs {
		addr, err := netip.ParseAddr(a.IP.String())
		if err != nil || !addr.IsValid() {
			continue
		}

		if ipv6 == addr.Is6() {
			return a
		}
	}
	return nil
}

func Interfaces() []*net.Interface {
	list := []*net.Interface{}

	ifis, err := net.Interfaces()
	if err != nil {
		return list
	}

	for _, ifi := range ifis {
		if !CheckInterface(&ifi) {
			continue
		}
		addrs := InterfaceAddrs(&ifi, nil)
		if addrs == nil {
			continue
		}
		addrStr := []string{}
		for _, addr := range addrs {
			addrStr = append(addrStr, addr.String())
		}
		ifiNew := ifi
		list = append(list, &ifiNew)
	}

	return list
}

func InterfaceByIndex(index int) *net.Interface {
	ifis := Interfaces()
	for _, ifi := range ifis {
		if ifi.Index == index {
			return ifi
		}
	}
	return nil
}

func InterfaceByName(name string) *net.Interface {
	if len(name) == 0 {
		return nil
	}
	ifis := Interfaces()
	for _, ifi := range ifis {
		if ifi.Name == name {
			return ifi
		}
	}
	return nil
}
