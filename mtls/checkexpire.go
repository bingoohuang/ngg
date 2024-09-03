package mtls

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"
)

// CheckSSLExpired 检查服务端证书过期情况，返回字符串不为空，说明证书即将过期
func CheckSSLExpired(url string) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			// 注意如果证书已过期，那么只有在关闭证书校验的情况下链接才能建立成功
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	for _, cert := range resp.TLS.PeerCertificates {
		if !cert.NotAfter.After(time.Now()) {
			return fmt.Sprintf("Website [%s] certificate has expired: %s", url,
				cert.NotAfter.Local().Format("2006-01-02 15:04:05")), nil
		}

		if cert.NotAfter.Sub(time.Now()) < 5*24*time.Hour {
			return fmt.Sprintf("Website [%s] certificate will expire, remaining time: %fh", url,
				cert.NotAfter.Sub(time.Now()).Hours()), nil
		}
	}
	log.Printf("the %s https no expired\n", url)

	return "", nil
}
