package gnet

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/bingoohuang/ngg/ss"
)

func GetMac() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return macAddrs
	}

	for _, i := range netInterfaces {
		if len(i.HardwareAddr) == 0 ||
			i.Flags&net.FlagUp != net.FlagUp ||
			i.Flags&net.FlagLoopback == net.FlagLoopback {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		found := false
	FOR:
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPAddr:
				if p4 := v.IP.To4(); len(p4) == net.IPv4len {
					found = true
					break FOR
				}
			case *net.IPNet:
				if p4 := v.IP.To4(); len(p4) == net.IPv4len {
					found = true
					break FOR
				}
			}
		}

		if found {
			mac := i.HardwareAddr.String()
			if !slices.Contains(macAddrs, mac) {
				macAddrs = append(macAddrs, mac)
			}
		}
	}

	return macAddrs
}

// IsIPv4 tells a string if in IPv4 format.
func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

// IsIPv6 tells a string if in IPv6 format.
func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

// ListIPv4 list all IPv4 addresses.
// ifaceNames are used to specified interface names (filename wild match pattern supported also, like eth*).
func ListIPv4(ifaceNames ...string) ([]string, error) {
	var ips []string

	f := func(ip net.IP) (yes bool) {
		s := ip.String()
		if IsIPv4(s) {
			ips = append(ips, s)
		}
		return true
	}
	_, err := ListIP(f, ifaceNames...)

	return ips, err
}

// ListIPv6 list all IPv6 addresses.
// ifaceNames are used to specified interface names (filename wild match pattern supported also, like eth*).
func ListIPv6(ifaceNames ...string) ([]string, error) {
	var ips []string

	f := func(ip net.IP) (yes bool) {
		s := ip.String()
		if IsIPv6(s) {
			ips = append(ips, s)
		}
		return true
	}
	_, err := ListIP(f, ifaceNames...)

	return ips, err
}

// ListIfaceNames list all net interface names.
func ListIfaceNames() (names []string) {
	list, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, i := range list {
		f := i.Flags
		if i.HardwareAddr == nil || f&net.FlagUp == 0 || f&net.FlagLoopback == 1 {
			continue
		}

		names = append(names, i.Name)
	}

	return names
}

// ListIP list all IP addresses.
func ListIP(predicate func(net.IP) bool, ifaceNames ...string) ([]net.IP, error) {
	list, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces, err: %w", err)
	}

	ips := make([]net.IP, 0)
	matcher := newIfaceNameMatcher(ifaceNames)

	for _, i := range list {
		f := i.Flags
		if i.HardwareAddr == nil ||
			f&net.FlagUp != net.FlagUp ||
			f&net.FlagLoopback == net.FlagLoopback ||
			!matcher.Matches(i.Name) {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		ips = collectAddresses(predicate, addrs, ips)
	}

	return ips, nil
}

func collectAddresses(predicate func(net.IP) bool, addrs []net.Addr, ips []net.IP) []net.IP {
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPAddr:
			ip = v.IP
		case *net.IPNet:
			ip = v.IP
		default:
			continue
		}

		if !ipContains(ips, ip) && predicate(ip) {
			ips = append(ips, ip)
		}
	}

	return ips
}

func ipContains(ips []net.IP, ip net.IP) bool {
	for _, j := range ips {
		if j.Equal(ip) {
			return true
		}
	}

	return false
}

// OutboundIP  gets preferred outbound ip of this machine.
func OutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}

	defer conn.Close()

	s := conn.LocalAddr().String()
	return s[:strings.LastIndex(s, ":")]
}

// MainIPv4 tries to get the main IP address and the IP addresses.
func MainIPv4(ifaceName ...string) (string, []string) {
	ips, _ := ListIPv4(ifaceName...)
	if len(ips) == 1 {
		return ips[0], ips
	}

	if s := findMainIPByIfconfig(ifaceName); s != "" {
		return s, ips
	}

	if out := OutboundIP(); out != "" && ss.AnyOf(out, ips...) {
		return out, ips
	}

	if len(ips) > 0 {
		return ips[0], ips
	}

	return "", nil
}

func findMainIPByIfconfig(ifaceName []string) string {
	names := ListIfaceNames()

	var matchedNames []string
	matcher := newIfaceNameMatcher(ifaceName)
	for _, n := range names {
		if matcher.Matches(n) {
			matchedNames = append(matchedNames, n)
		}
	}

	if len(matchedNames) == 0 {
		return ""
	}

	name := matchedNames[0]
	for _, n := range matchedNames {
		// for en0 on mac or eth0 on linux
		if strings.HasPrefix(n, "e") && strings.HasSuffix(n, "0") {
			name = n
			break
		}
	}

	/*
		[root@tencent-beta17 ~]# ifconfig eth0
		eth0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
		        inet 192.168.108.7  netmask 255.255.255.0  broadcast 192.168.108.255
		        ether 52:54:00:ef:16:bd  txqueuelen 1000  (Ethernet)
		        RX packets 1838617728  bytes 885519190162 (824.7 GiB)
		        RX errors 0  dropped 0  overruns 0  frame 0
		        TX packets 1665532349  bytes 808544539610 (753.0 GiB)
		        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
	*/
	re := regexp.MustCompile(`inet\s+([\w.]+?)\s+`)
	c := exec.Command("ifconfig", name)
	if co, err := c.Output(); err == nil {
		sub := re.FindStringSubmatch(string(co))
		if len(sub) > 1 {
			return sub[1]
		}
	}

	return ""
}

type ifaceNameMatcher struct {
	ifacePatterns map[string]bool
}

func newIfaceNameMatcher(names []string) ifaceNameMatcher {
	return ifaceNameMatcher{ifacePatterns: ss.ToSet(names)}
}

func (i ifaceNameMatcher) Matches(name string) bool {
	if len(i.ifacePatterns) == 0 {
		return true
	}
	if i.ifacePatterns[name] {
		return true
	}

	for k := range i.ifacePatterns {
		if ok, _ := ss.FnMatch(k, name, true); ok {
			return true
		}
	}

	for k := range i.ifacePatterns {
		if strings.Contains(k, name) {
			return true
		}
	}

	return false
}
