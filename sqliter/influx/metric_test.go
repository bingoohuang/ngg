package influx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	lp := "cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000"
	point, err := ParseLineProtocol(lp)
	assert.Nil(t, err)

	assert.Equal(t, "cpu_load_short", point.Name())
	assert.Equal(t, map[string]string{
		"host":   "server01",
		"region": "us-west",
	}, point.Tags())
	assert.Equal(t, map[string]any{
		"value": 0.64,
	}, point.Fields())
	assert.Equal(t, "2015-06-12T04:46:02+08:00", point.Time().Format(time.RFC3339))

	p, err := ParseLineProtocol(`net,host=localhost.localdomain,interface=all,ip=192.168.112.71,ips=192.168.112.71 tcp_incsumerrors=0i,ip_reasmoks=0i,ip_forwdatagrams=0i,icmpmsg_outtype0=495i,icmp_inechoreps=7i,tcp_rtomin=200i,ip_inaddrerrors=0i,icmp_inmsgs=2082i,icmp_intimeexcds=0i,ip_reasmreqds=0i,icmpmsg_intype8=495i,ip_indiscards=0i,ip_inhdrerrors=0i,tcp_rtomax=120000i,ip_outdiscards=0i,tcp_activeopens=879152i,tcp_attemptfails=870578i,icmp_outechoreps=495i,icmp_outerrors=0i,icmp_outechos=7i,icmp_outsrcquenchs=0i,ip_forwarding=2i,ip_fragcreates=0i,ip_outnoroutes=0i,icmpmsg_outtype3=535i,icmp_inechos=495i,icmp_intimestamps=0i,icmp_incsumerrors=0i,udplite_inerrors=0i,icmpmsg_intype0=7i,tcp_insegs=1325208840i,ip_indelivers=1331787493i,icmp_outaddrmaskreps=0i,tcp_rtoalgorithm=1i,icmp_intimestampreps=0i,tcp_currestab=11i,ip_fragoks=0i,ip_fragfails=0i,ip_reasmfails=0i,icmp_indestunreachs=1580i,tcp_outsegs=1405660253i,udplite_noports=0i,udplite_incsumerrors=0i,ip_reasmtimeout=0i,udp_outdatagrams=3785i,ip_defaultttl=64i,icmp_inerrors=0i,udp_inerrors=0i,udp_rcvbuferrors=0i,udplite_sndbuferrors=0i,icmpmsg_intype3=1580i,udp_incsumerrors=0i,icmp_insrcquenchs=0i,icmp_inaddrmaskreps=0i,udp_sndbuferrors=0i,icmp_inparmprobs=0i,tcp_retranssegs=129904i,icmp_outparmprobs=0i,udplite_rcvbuferrors=0i,icmp_inaddrmasks=0i,icmp_outtimestamps=0i,icmpmsg_outtype8=7i,ip_inunknownprotos=0i,udplite_outdatagrams=0i,tcp_estabresets=355i,icmp_outmsgs=1037i,icmp_outdestunreachs=535i,udplite_indatagrams=0i,udp_noports=712i,tcp_outrsts=1167i,icmp_outtimestampreps=0i,ip_outrequests=1402096376i,icmp_outtimeexcds=0i,icmp_outredirects=0i,tcp_passiveopens=649i,icmp_outaddrmasks=0i,tcp_maxconn=-1i,tcp_inerrs=0i,icmp_inredirects=0i,ip_inreceives=1331787725i,udp_indatagrams=0i 1723795322000000000`)
	assert.Nil(t, err)
	t.Logf("point: %+v", p)
}
