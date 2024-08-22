package gnet_test

import (
	"testing"

	"github.com/bingoohuang/ngg/gnet"
	"github.com/stretchr/testify/assert"
)

func TestListAllIPv4(t *testing.T) {
	ips, err := gnet.ListIPv4()
	assert.Nil(t, err)
	t.Logf("ListIPv4: %+v", ips)

	ips, err = gnet.ListIPv6()
	assert.Nil(t, err)
	t.Logf("ListIPv6: %+v", ips)
}

func TestGetOutboundIP(t *testing.T) {
	t.Logf("Outbound: %s", gnet.OutboundIP())
	mainIP, ipList := gnet.MainIPv4()
	t.Logf("MainIP: %s, ipList: %+v", mainIP, ipList)

	t.Logf("Mac: %+v", gnet.GetMac())
}
