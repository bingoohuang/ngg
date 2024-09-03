package mtls

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/grantae/certinfo"
)

func CertInfo(certFile string) error {
	// Read and parse the PEM certificate file
	pemData, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("read %s: %w", certFile, err)
	}
	block, rest := pem.Decode(pemData)
	if block == nil || len(rest) > 0 {
		return fmt.Errorf("certificate decoding error")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("x509.ParseCertificate: %w", err)
	}

	// Print the certificate
	result, err := certinfo.CertificateText(cert)
	if err != nil {
		return fmt.Errorf("certinfo.CertificateText: %w", err)
	}
	fmt.Print(result)

	return nil
}
