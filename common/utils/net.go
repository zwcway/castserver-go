package utils

import (
	"net"
	"net/netip"
	"strings"

	"github.com/jackpal/gateway"
)

func IsConnectCloseError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "use of closed network connection")
}

func MacIsValid(m net.HardwareAddr) bool {
	return m[0]+m[1]+m[2]+m[3]+m[4]+m[5] != 0
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

// 默认可连外网的ip
func DefaultAddr() *netip.Addr {
	ip, err := gateway.DiscoverInterface()
	if err == nil {
		addr, err := netip.ParseAddr(ip.String())
		if err != nil {
			return nil
		}
		return &addr
	}

	ifis, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for i := 0; i < len(ifis); i++ {
		ifi := &ifis[i]
		if !CheckInterface(ifi) {
			continue
		}
		addrs := InterfaceAddrs(ifi, nil)
		if addrs == nil {
			continue
		}
		for _, addr := range addrs {
			return IpNetToAddr(addr)
		}
	}

	return nil
}

func DefaultInterface() *net.Interface {
	ifis, err := net.Interfaces()
	if err != nil {
		return nil
	}
	addr := DefaultAddr()
	if addr == nil {
		return nil
	}
	for i := 0; i < len(ifis); i++ {
		ifi := &ifis[i]
		if !CheckInterface(ifi) {
			continue
		}
		addrs := InterfaceAddrs(ifi, nil)
		if addrs == nil {
			continue
		}
		for _, addr := range addrs {
			if addr.IP.Equal(addr.IP) {
				return ifi
			}
		}
	}
	return nil
}

func IpNetToAddr(addr *net.IPNet) *netip.Addr {
	ip := addr.IP
	if len(addr.Mask) == net.IPv4len {
		ip = ip[len(ip)-4:]
	}

	if ip, ok := netip.AddrFromSlice(ip); ok {
		return &ip
	}
	return nil
}

func Interfaces() []*net.Interface {
	list := []*net.Interface{}

	ifis, err := net.Interfaces()
	if err != nil {
		return list
	}

	for i := 0; i < len(ifis); i++ {
		ifi := &ifis[i]
		if !CheckInterface(ifi) {
			continue
		}
		addrs := InterfaceAddrs(ifi, nil)
		if addrs == nil {
			continue
		}
		// addrStr := make([]string, len(addrs))
		// for i, addr := range addrs {
		// 	addrStr[i] = addr.String()
		// }
		list = append(list, ifi)
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
