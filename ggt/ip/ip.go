//go:build all || ip

package ip

import (
	"fmt"
	"log"
	"net"

	"github.com/atotto/clipboard"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gnet"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	Register(root.Cmd)
}

func (f *Cmd) run(args []string) error {
	if !f.v4 && !f.v6 {
		f.v4 = true
	}
	var ifaceNames []string
	if f.iface != "" {
		ifaceNames = append(ifaceNames, f.iface)
	}

	mainIP, ipList := gnet.MainIPv4(ifaceNames...)
	if len(ipList) > 1 {
		log.Printf("IP: %v", ipList)
	}
	log.Printf("Main IP: %s", mainIP)
	log.Printf("Outbound IP: %v", gnet.OutboundIP())

	if f.v4 {
		allIPv4, _ := gnet.ListIPv4(ifaceNames...)
		log.Printf("IPv4: %v", allIPv4)
	}

	if f.v6 {
		allIPv6, _ := gnet.ListIPv6(ifaceNames...)
		log.Printf("IPv6: %v", allIPv6)
	}

	log.Printf("Mac addresses: %v", gnet.GetMac())

	if publicIP, err := StunPublicIP(f.stun); err != nil {
		log.Printf("stun error: %v", err)
	} else if len(publicIP) > 0 {
		log.Printf("Stun IP: %v ✅", publicIP)
		clipboard.WriteAll(publicIP[0])
	}

	if f.verbose {
		ListIfaces(f.v4, f.v6, f.iface)
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

func (f *Cmd) initFlags(p *pflag.FlagSet) {
	p.BoolVarP(&f.v4, "v4", "4", false, "only show ipv4")
	p.BoolVarP(&f.v6, "v6", "6", false, "only show ipv6")
	p.BoolVarP(&f.verbose, "verbose", "v", false, "Verbose output for more details")
	p.StringVarP(&f.iface, "iface", "i", "", "Interface name pattern specified(eg. eth0, eth*)")
	p.StringSliceVar(&f.stun, "stun", defaultStunServers, "STUN server")
}

type Cmd struct {
	*root.RootCmd
	iface   string
	stun    []string
	verbose bool
	v4, v6  bool
}

func Register(rootCmd *root.RootCmd) {
	c := &cobra.Command{
		Use:   "ip",
		Short: "show host IP addresses",
	}

	fc := &Cmd{RootCmd: rootCmd}
	c.Run = func(cmd *cobra.Command, args []string) {
		if err := fc.run(args); err != nil {
			fmt.Println(err)
		}
	}
	fc.initFlags(c.Flags())
	rootCmd.AddCommand(c)
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
