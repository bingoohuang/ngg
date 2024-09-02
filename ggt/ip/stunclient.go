//go:build all || ip

package ip

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/pion/stun/v2"
)

func StunPublicIP(stunServers []string) ([]string, error) {
	var publicIPs sync.Map
	var wg sync.WaitGroup
	portPostfix := regexp.MustCompile(`.+:\d+$`)
	for _, addr := range stunServers {
		if !portPostfix.MatchString(addr) {
			addr += ":3478"
		}
		if !strings.HasPrefix(addr, "stun:") {
			addr = "stun:" + addr
		}

		wg.Add(1)
		go func(stunServer string) {
			defer wg.Done()

			if publicIP, err := StunAddr(stunServer); err != nil {
				log.Printf("Stun %s, error: %v", stunServer, err)
			} else if publicIP != "" {
				// log.Printf("Stun %s, PublicIP: %v", stunServer, publicIP)
				publicIPs.Store(publicIP, true)
			}
		}(addr)
	}

	wg.Wait()
	var result []string
	publicIPs.Range(func(key, value any) bool {
		result = append(result, key.(string))
		return true
	})
	return result, nil
}

func StunAddr(uriStr string) (string, error) {
	uri, err := stun.ParseURI(uriStr)
	if err != nil {
		return "", fmt.Errorf("invalid URI %q: %w", uriStr, err)
	}

	// we only try the first address, so restrict ourselves to IPv4
	c, err := stun.DialURI(uri, &stun.DialConfig{})
	if err != nil {
		return "", fmt.Errorf("dial %s: %w", uri, err)
	}

	var xorAddr stun.XORMappedAddress

	if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
		if res.Error != nil {
			log.Printf("STUN transaction error: %s", res.Error)
			return
		}

		if err := xorAddr.GetFrom(res.Message); err != nil {
			log.Printf("get %s XOR-MAPPED-ADDRESS error: %v", uriStr, err)
			return
		}

	}); err != nil {
		return "", fmt.Errorf("stun Do error: %w", err)
	}
	if err := c.Close(); err != nil {
		log.Fatalf("Failed to close connection: %s", err)
	}

	if len(xorAddr.IP) == 0 {
		return "", nil
	}

	return xorAddr.IP.String(), nil
}
