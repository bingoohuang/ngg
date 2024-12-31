//go:build !android

package ggtip

import (
	"log"

	"github.com/bingoohuang/ngg/gnet"
)

func stunOthers(stunServer []string) {
	if publicIP, err := StunPublicIP(stunServer); err != nil {
		log.Printf("stun error: %v", err)
	} else if len(publicIP) > 0 {
		log.Printf("Stun IP: %v ✅", publicIP)

		for _, public := range publicIP {
			if gnet.IsIPv4(public) {
				// clipboard.WriteAll(public)
				// log.Printf("%s copied to clipboard", public)
				break
			}
		}
	}
}
