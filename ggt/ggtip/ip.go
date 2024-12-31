package ggtip

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/spf13/cobra"
)

func Run() {
	c := root.CreateCmd(nil, "ip", "show host IP addresses", &subCmd{})
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type subCmd struct {
	Iface      string   `short:"i" help:"Interface name pattern specified(eg. eth0, eth*)"`
	StunServer []string `help:"STUN server"`
	Stun       bool     `help:"Check IP by STUN servers"`
	Verbose    bool     `short:"v" help:"Verbose output for more details"`
	V4         bool     `short:"4" help:"only show ipv4"`
	V6         bool     `short:"6" help:"only show ipv6"`
	ListIfaces bool     `short:"l" help:"list all interfaces"`
}

func (f *subCmd) Run(*cobra.Command, []string) error {
	if !f.V4 && !f.V6 {
		f.V4 = true
	}
	var ifaceNames []string
	if f.Iface != "" {
		ifaceNames = append(ifaceNames, f.Iface)
	}

	mainIP, ipList := gnet.MainIPv4(ifaceNames...)
	if len(ipList) > 1 {
		log.Printf("IP: %v", ipList)
	}
	log.Printf("Main IP: %s", mainIP)
	log.Printf("Outbound IP: %v", gnet.OutboundIP())

	if f.V4 {
		allIPv4, _ := gnet.ListIPv4(ifaceNames...)
		log.Printf("IPv4: %v", allIPv4)
	}

	if f.V6 {
		allIPv6, _ := gnet.ListIPv6(ifaceNames...)
		log.Printf("IPv6: %v", allIPv6)
	}

	log.Printf("Mac addresses: %v", gnet.GetMac())

	if f.Stun {
		stunOthers(f.StunServer)
	}

	if f.ListIfaces {
		ListIfaces(f.V4, f.V6, f.Iface)
	}

	if f.Verbose {
		CheckPublicIP()
	}

	return nil
}

// https://github.com/dndx/uip
// public-stun-list.txt  https://gist.github.com/mondain/b0ec1cf5f60ae726202e
var defaultStunServers = []string{
	"stun.l.google.com:19302",
	"stun.cloudflare.com",
	"stun.syncthing.net",
	"stun.miwifi.com",
	"stun.chat.bilibili.com",
}

func (f *subCmd) DefaultPlagValues(name string) (any, bool) {
	switch name {
	case "StunServer":
		return defaultStunServers, true
	}
	return nil, false
}

// ListIfaces 根据mode 列出本机所有IP和网卡名称.
func ListIfaces(v4, v6 bool, ifaceName string) {
	list, err := net.Interfaces()
	if err != nil {
		log.Printf("failed to get interfaces, err: %v", err)
		return
	}

	for _, f := range list {
		listIface(f, v4, v6)
	}
}

func listIface(f net.Interface, v4, v6 bool) {
	if f.HardwareAddr == nil || f.Flags&net.FlagUp == 0 || f.Flags&net.FlagLoopback == 1 {
		return
	}

	addrs, err := f.Addrs()
	if err != nil {
		log.Printf("\t failed to f.Addrs, × err: %v", err)
		return
	}

	if len(addrs) == 0 {
		return
	}

	got := false
	for _, a := range addrs {
		var netip net.IP
		log.Printf("addr(%T): %s", a, a)
		switch v := a.(type) {
		case *net.IPAddr:
			netip = v.IP
		case *net.IPNet:
			netip = v.IP
		default:
			log.Print("\t\t not .(*net.IPNet) or .(*net.IPNet) ×")
			continue
		}

		if gnet.IsIPv4(netip.String()) && !v4 || gnet.IsIPv6(netip.String()) && !v6 {
			continue
		}

		if netip.IsLoopback() {
			log.Print("\t\t IsLoopback ×")
			continue
		}

		got = true
	}

	mac := f.HardwareAddr.String()

	if got {
		log.Printf("\tmac: %s, addrs %+v √ ✅", mac, addrs)
	} else {
		log.Printf("\tmac: %s, addrs %+v ×", mac, addrs)
	}
}
