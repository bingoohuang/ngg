package ggtip

import (
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/gnet/android"
	"github.com/imroc/req/v3"
)

var Endpoints = []string{
	"https://api.maao.cc/ip/",
	"https://d5k.top/ping",
	"https://api.ipify.org?format=json",
	"https://alanwei.azurewebsites.net/api/tool/ip",
	"https://httpbin.org/ip",
	"https://ident.me",
	"ip.gs",
	"ip.sb",
	"cip.cc",
	"icanhazip.com",
	"ipv4.icanhazip.com",
	"api.ipify.org",
	"ifconfig.me",
	"ifconfig.co",
	"ipecho.net/plain",
	"whatismyip.akamai.com",
	"inet-ip.info",
	"myip.ipip.net",
	"ipinfo.io",
	"ifcfg.cn",
	"4.ipw.cn",
	"members.3322.org/dyndns/getip",
	"curlmyip.com",
	"ip.appspot.com",
	"www.trackip.net/ip",

	"https://myexternalip.com/raw",
	"http://checkip.amazonaws.com",
	"http://myip.dnsomatic.com/",
	"http://diagnostic.opendns.com/myip",
}

func init() {
	for i, ipUrl := range Endpoints {
		if !strings.HasPrefix(ipUrl, "http") {
			Endpoints[i] = "http://" + ipUrl
		}
	}
}

func CheckPublicIP() {
	ipv4ch := make(chan []string)
	for _, ipUrl := range Endpoints {
		log.Printf("Checking %s", ipUrl)
		go invoke(ipUrl, ipv4ch)
	}

	ipv4Map := map[string]int{}
	for range Endpoints {
		for _, ip := range <-ipv4ch {
			ipv4Map[ip]++
		}
	}
	most := 0
	for _, v := range ipv4Map {
		most = max(v, most)
	}

	for k, v := range ipv4Map {
		if most == v {
			log.Printf("%d/%d Public IP: %s ✅", v, len(Endpoints), k)
			//  clipboard.WriteAll(k)
		} else {
			log.Printf("%d/%d Public IP: %s", v, len(Endpoints), k)
		}
	}
}

var (
	cutBlanks = regexp.MustCompile(`\s+`)
	ipv4Reg   = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
)

func invoke(ipUrl string, ipv4ch chan<- []string) {
	var ipv4s []string
	defer func() {
		ipv4ch <- ipv4s
	}()

	start := time.Now()
	if res, err := client.R().
		SetHeader("User-Agent", "curl").
		Get(ipUrl); err == nil {
		if data := res.Bytes(); len(data) > 0 {
			ipv4s = ipv4Reg.FindAllString(string(data), -1)
			data = cutBlanks.ReplaceAll(data, []byte(" "))
			if len(ipv4s) == 1 && ipv4s[0] == string(data) {
				log.Printf("%v\t%s\t%s", ipv4s, FormatDuration(time.Since(start)), ipUrl)
			} else {
				log.Printf("%v\t%s\t%s\t%s", ipv4s, FormatDuration(time.Since(start)), ipUrl, data)
			}
		}
	} else {
		log.Printf("check %s error: %v", ipUrl, err)
	}
}

var client = req.C().
	SetTimeout(5 * time.Second).
	SetDial((&net.Dialer{Resolver: android.Resolver}).DialContext).
	SetProxy(nil) // Disable proxy

// FormatDuration formats a duration with a precision of 3 digits
// if it is less than 100s.
// https://stackoverflow.com/a/68870075
func FormatDuration(d time.Duration) time.Duration {
	scale := 100 * time.Second
	// look for the max scale that is smaller than d
	for scale > d {
		scale = scale / 10
	}
	return d.Round(scale / 100)
}
