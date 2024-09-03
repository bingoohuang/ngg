// Package mtls demonstrate Mutual TLS usage.
package mtls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// MakeCerts make certs.
// 环境变量 DNS_NAMES 指定 DNS 域名
func MakeCerts(path string, clientNames []string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %s: %v", path, err)
	}

	subject := pkix.Name{
		Country:            []string{"Earth"},
		Organization:       []string{"CA Company"},
		OrganizationalUnit: []string{"Engineering"},
		Locality:           []string{"Mountain"},
		Province:           []string{"Asia"},
		StreetAddress:      []string{"Bridge"},
		PostalCode:         []string{"123456"},
		SerialNumber:       "",
		CommonName:         "CA",
		Names:              []pkix.AttributeTypeAndValue{},
		ExtraNames:         []pkix.AttributeTypeAndValue{},
	}
	keyPair, err := makeRootCA(path, &subject)
	if err != nil {
		return fmt.Errorf("make CA Certificate: %w", err)
	}
	log.Printf("Created ✅. CertFile: %s, KeyFile: %s",
		keyPair.CertFile, keyPair.KeyFile)

	for _, name := range append([]string{"server"}, clientNames...) {
		subject.CommonName = name
		subject.Organization = []string{name + " Company"}

		filePair, err := makeCert(keyPair, &subject, path, name)
		if err != nil {
			return fmt.Errorf("make %s Certificate: %w", name, err)
		}
		log.Printf("Created and Signed %s ✅. CertFile: %s, KeyFile: %s",
			name, filePair.CertFile, filePair.KeyFile)
	}

	return nil
}

func fileExists(file string) bool {
	// Check if file exists
	stat, err := os.Stat(file)
	return err == nil && !stat.IsDir()
}

func makeRootCA(path string, subject *pkix.Name) (*KeyPair, error) {
	caCrtPath := filepath.Join(path, "ca.crt")
	caKeyPath := filepath.Join(path, "ca.key")
	if fileExists(caCrtPath) && fileExists(caKeyPath) {
		return ReadCertificateAuthority(caCrtPath, caKeyPath)
	}

	// 生成证书的序列号
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("rand serialNumber: %w", err)
	}

	// creating a CA which will be used to sign all of our certificates using the x509 package from the Go Standard Library
	caCert := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               *subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10*365, 0, 0),
		IsCA:                  true, // <- indicating this certificate is a CA certificate.
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	// generate a private key for the CA
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate the CA Private Key: %w", err)
	}

	// create the CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("create the CA Certificate: %w", err)
	}

	// Create the CA PEM files
	caPEM := new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		return nil, fmt.Errorf("pem encode: %w", err)
	}

	if err := os.WriteFile(caCrtPath, caPEM.Bytes(), 0o644); err != nil {
		return nil, fmt.Errorf("write the CA certificate file: %w", err)
	}

	keyPEM := new(bytes.Buffer)
	if err := pem.Encode(keyPEM, &pem.Block{Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey)}); err != nil {
		return nil, fmt.Errorf("pem encode: %w", err)
	}

	if err := os.WriteFile(caKeyPath, keyPEM.Bytes(), 0o644); err != nil {
		return nil, fmt.Errorf("write the CA certificate file: %w", err)
	}

	return &KeyPair{Cert: caCert, Key: caKey,
		FilePair: FilePair{CertFile: caCrtPath, KeyFile: caKeyPath}}, nil
}

func makeCert(keyPair *KeyPair, subject *pkix.Name, path, name string) (*FilePair, error) {
	crtFile := filepath.Join(path, name+".crt")
	keyFile := filepath.Join(path, name+".key")
	filePair := &FilePair{CertFile: crtFile, KeyFile: keyFile}
	if fileExists(crtFile) && fileExists(keyFile) {
		return filePair, nil
	}

	// 生成证书的序列号
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("rand serialNumber: %w", err)
	}

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      *subject,
		// IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     GetEnvStringSlice("DNS_NAMES", []string{"localhost", "d5k.co", "d5k.top"}),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Printf("Generate the Key error: %v\n", err)
		return nil, err
	}
	certBytes, err := x509.CreateCertificate(rand.Reader,
		cert, keyPair.Cert, &certKey.PublicKey, keyPair.Key)
	if err != nil {
		log.Printf("Generate the certificate error: %v\n", err)
		return nil, err
	}

	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return nil, fmt.Errorf("pem: %w", err)
	}

	if err := os.WriteFile(crtFile, certPEM.Bytes(), 0o644); err != nil {
		log.Printf("Write the CA certificate file error: %v\n", err)
		return nil, err
	}

	certKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(certKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certKey),
	}); err != nil {
		return nil, fmt.Errorf("pem: %w", err)
	}
	if err := os.WriteFile(keyFile, certKeyPEM.Bytes(), 0o644); err != nil {
		log.Printf("Write the CA certificate file error: %v\n", err)
		return nil, err
	}

	return filePair, nil
}
