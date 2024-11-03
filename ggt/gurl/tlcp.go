package gurl

import (
	"fmt"
	"net"
	"os"
	"strings"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/bingoohuang/ngg/ggt/gurl/certinfo"
	"github.com/bingoohuang/ngg/ss"
	"github.com/emmansun/gmsm/smx509"
)

var tlcpSessionCache tlcp.SessionCache

func init() {
	if cacheSize, _ := ss.Getenv[int](`TLCP_SESSION_CACHE`, 32); cacheSize > 0 {
		tlcpSessionCache = tlcp.NewLRUSessionCache(cacheSize)
	}
}

var tlcpCerts = os.Getenv("TLCP_CERTS")

func createTlcpDialer(dialer *net.Dialer, caFile string) DialContextFn {
	c := &tlcp.Config{
		InsecureSkipVerify: !ss.Pick1(ss.GetenvBool(`TLS_VERIFY`, false)),
		SessionCache:       tlcpSessionCache,
	}

	c.EnableDebug = HasPrintOption(printDebug)

	if caFile != "" {
		rootCert, err := smx509.ParseCertificatePEM(ss.Must(ss.ReadFile(caFile)))
		if err != nil {
			panic(err)
		}
		pool := smx509.NewCertPool()
		pool.AddCert(rootCert)
		c.RootCAs = pool
	}

	if tlcpCerts != "" {
		// TLCP 1.1，套件ECDHE-SM2-SM4-CBC-SM3，设置客户端双证书
		certsFiles := strings.Split(tlcpCerts, ",")
		var certs []tlcp.Certificate
		switch len(certsFiles) {
		case 0, 2, 4:
		default:
			panic("$TLCP_CERTS should be sign.cert.pem,sign.key.pem,enc.cert.pem,enc.key.pem")
		}
		if len(certsFiles) >= 2 {
			signCertKeypair, err := tlcp.X509KeyPair(ss.Must(ss.ReadFile(certsFiles[0])), ss.Must(ss.ReadFile(certsFiles[1])))
			if err != nil {
				panic(err)
			}
			certs = append(certs, signCertKeypair)
		}
		if len(certsFiles) >= 4 {
			encCertKeypair, err := tlcp.X509KeyPair(ss.Must(ss.ReadFile(certsFiles[2])), ss.Must(ss.ReadFile(certsFiles[3])))
			if err != nil {
				panic(err)
			}
			certs = append(certs, encCertKeypair)
		}

		if len(certs) > 0 {
			c.Certificates = certs
			c.CipherSuites = []uint16{tlcp.ECDHE_SM4_CBC_SM3, tlcp.ECDHE_SM4_GCM_SM3}
		}
	}

	if c.EnableDebug {
		fmt.Printf("load %d client certs\n", len(c.Certificates))
	}

	d := tlcp.Dialer{NetDialer: dialer, Config: c}
	return d.DialContext
}

func printTLCPConnectState(conn net.Conn, state tlcp.ConnectionState) {
	if !HasPrintOption(printRspOption) {
		return
	}

	fmt.Printf("option Conn type: %T\n", conn)
	fmt.Printf("option TLCP.Version: %s\n", func(version uint16) string {
		switch version {
		case tlcp.VersionTLCP:
			return "TLCP"
		default:
			return "Unknown"
		}
	}(state.Version))
	for i, cert := range state.PeerCertificates {
		text, _ := certinfo.CertificateText(cert.ToX509())
		fmt.Printf("option Cert[%d]: %s\n", i, text)
	}
	fmt.Printf("option TLCP.HandshakeComplete: %t\n", state.HandshakeComplete)
	fmt.Printf("option TLCP.DidResume: %t\n", state.DidResume)
	for _, suit := range tlcp.CipherSuites() {
		if suit.ID == state.CipherSuite {
			fmt.Printf("option TLCP.CipherSuite: %+v", suit)
			break
		}
	}
	fmt.Println()
}
