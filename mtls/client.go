package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetEnvStringSlice get bool values from an env.
func GetEnvStringSlice(name string, defaultValue []string) []string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}

	return strings.Split(value, ",")
}

func If[T any](condition bool, a, b T) T {
	if condition {
		return a
	}

	return b
}

// GetEnvBool get bool values from an env.
func GetEnvBool(name string) bool {
	switch strings.ToLower(os.Getenv(name)) {
	case "y", "yes", "on", "ok", "t", "true", "1":
		return true
	}
	return false
}

// StartClient start a https client invoke using the clientName's certs.
func StartClient(certsPath, clientName, urlAddress string) error {
	tlcConfig := &tls.Config{
		InsecureSkipVerify: GetEnvBool("INSECURE_SKIP_VERIFY"),
	}

	if !GetEnvBool("CLIENT_CA_OFF") {
		cert, err := os.ReadFile(filepath.Join(certsPath, "ca.crt"))
		if err != nil {
			return fmt.Errorf("open certificate file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(cert)

		clientCert := filepath.Join(certsPath, clientName+".crt")
		clientKey := filepath.Join(certsPath, clientName+".key")
		log.Println("Load key pairs - ", clientCert, clientKey)
		certificate, err := tls.LoadX509KeyPair(clientCert, clientKey)
		if err != nil {
			return fmt.Errorf("load certificate: %w", err)
		}

		tlcConfig.RootCAs = caCertPool
		tlcConfig.Certificates = []tls.Certificate{certificate}
	}

	client := http.Client{
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: tlcConfig,
		},
	}

	// Request /hello over port 8443 via the GET method
	// Using curl to verify it :
	// curl --trace trace.log -k --cacert certs/ca.crt  --cert certs/client.crt --key certs/client.key https://d5k.co:8443

	r, err := client.Get(urlAddress)
	if err != nil {
		return fmt.Errorf("making get request: %w", err)
	}

	// Read the response body
	defer func(r io.ReadCloser) {
		_ = r.Close()
	}(r.Body)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	PrintTLSConnectionState(r.TLS)

	// Print the response body to stdout
	log.Printf("%s", body)
	return nil
}
