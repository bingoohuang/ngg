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
	"time"

	"go.uber.org/multierr"
)

// [Implementing mTLS in Go](https://ayada.dev/posts/implementing-mtls-in-go/)

// Verify function will verify the client and server certificates using the root certificate that was provided in the function arguments.
// If generateCertificate is set to true, it will also generate new client and server certificates that is signed by root certificate.
func Verify(serverName, clientCert, clientKey, serverCert, serverKey, rootCa, rootKey string, generateCertificate bool) error {
	ca, err := ReadCertificateAuthority(rootCa, rootKey)
	if err != nil {
		return fmt.Errorf("read ca certificate: %w", err)
	}

	if generateCertificate {
		// generate and sign client certificate using root certificate.
		if err := GenerateAndSignCertificate(ca, clientCert, clientKey); err != nil {
			return fmt.Errorf("generate client certificate: %w", err)
		}

		// generate and sign server certificate using root certificate.
		if err := GenerateAndSignCertificate(ca, serverCert, serverKey); err != nil {
			return fmt.Errorf("generate server certificate: %w", err)
		}
	}

	cCert, err := ReadCertificate(clientCert, clientKey)
	if err != nil {
		return fmt.Errorf("read certificate: %w", err)
	}

	sCert, err := ReadCertificate(serverCert, serverKey)
	if err != nil {
		return fmt.Errorf("to read certificate: %w", err)
	}

	roots := x509.NewCertPool()
	roots.AddCert(ca.Cert)
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: x509.NewCertPool(),
		DNSName:       serverName,
	}

	// verify client certificate; return err on failure.
	if _, err := cCert.Cert.Verify(opts); err != nil {
		return fmt.Errorf("verify client certificate: %w", err)
	}

	// verify server certificate; return err on failure.
	if _, err := sCert.Cert.Verify(opts); err != nil {
		return fmt.Errorf("verify server certificate: %w", err)
	}

	log.Print("client and server cert verification succeeded")
	return nil
}

// ReadCertificate reads and parses the certificates from files provided as argument to this function.
func ReadCertificate(publicKeyFile, privateKeyFile string) (*KeyPair, error) {
	cert := new(KeyPair)

	privKey, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	privPemBlock, _ := pem.Decode(privKey)

	// Note that we use PKCS1 to parse the private key here.
	parsedPrivKey, errParse := x509.ParsePKCS1PrivateKey(privPemBlock.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("parse private key: %w", errParse)
	}

	cert.Key = parsedPrivKey

	pubKey, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	publicPemBlock, _ := pem.Decode(pubKey)

	parsedPubKey, errParse := x509.ParseCertificate(publicPemBlock.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("parse public key: %w", errParse)
	}

	cert.Cert = parsedPubKey

	return cert, nil
}

// ReadCertificateAuthority reads and parses the root certificate from files provided as argument to this function.
func ReadCertificateAuthority(publicKeyFile, privateKeyFile string) (*KeyPair, error) {
	root := new(KeyPair)

	rootKey, errRead := os.ReadFile(privateKeyFile)
	if errRead != nil {
		return nil, fmt.Errorf("read private key: %w", errRead)
	}

	privPemBlock, _ := pem.Decode(rootKey)
	if privPemBlock == nil {
		return nil, fmt.Errorf("pem decode %s failed", rootKey)
	}

	var err error
	root.Key, err = parsePrivateKey(privPemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	rootCert, errRead := os.ReadFile(publicKeyFile)
	if errRead != nil {
		return nil, fmt.Errorf("read public key: %w", errRead)
	}

	publicPemBlock, _ := pem.Decode(rootCert)
	if publicPemBlock == nil {
		return nil, fmt.Errorf("pem decode %s failed", rootCert)
	}

	rootPubCrt, errParse := x509.ParseCertificate(publicPemBlock.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("parse public key: %w", errParse)
	}

	root.Cert = rootPubCrt

	return root, nil
}

func parsePrivateKey(privPemBlockBytes []byte) (*rsa.PrivateKey, error) {
	// Note that we use PKCS8 to parse the private key here.
	k1, err1 := x509.ParsePKCS8PrivateKey(privPemBlockBytes)
	if err1 == nil {
		return k1.(*rsa.PrivateKey), nil
	}

	k2, err2 := x509.ParsePKCS1PrivateKey(privPemBlockBytes)
	if err2 == nil {
		return k2, nil
	}

	return nil, fmt.Errorf("parse private key: %w", multierr.Append(err1, err2))
}

// GenerateAndSignCertificate method will use the root certificate's public and private key to generate a certificate and sign it.
// The certificate's public and private keys will be stored in the files provided as argument to this function.
// openssl req -newkey rsa:2048 -nodes -x509 -days 3650 -out certs/ca.crt  -keyout certs/ca.key -subj "/C=US/ST=California/L=San Francisco/O=ayada/OU=dev/CN=localhost"
func GenerateAndSignCertificate(root *KeyPair, publicKeyFile, privateKeyFile string) error {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"ayada"},
			Country:      []string{"US"},
			Province:     []string{"California"},
			Locality:     []string{"San Francisco"},
			CommonName:   "localhost",
		},
		DNSNames: GetEnvStringSlice("DNS_NAMES", []string{"localhost", "d5k.co"}),
		// IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, root.Cert, &privKey.PublicKey, root.Key)
	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return fmt.Errorf("pem: %w", err)
	}
	certKey := new(bytes.Buffer)
	if err := pem.Encode(certKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)}); err != nil {
		return fmt.Errorf("pem: %w", err)
	}

	if err := os.WriteFile(publicKeyFile, certPEM.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write to CERTIFICATE file: %w", err)
	}

	if err := os.WriteFile(privateKeyFile, certKey.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write to private key file: %w", err)
	}

	return nil
}

type FilePair struct {
	CertFile string
	KeyFile  string
}

type KeyPair struct {
	Cert *x509.Certificate
	Key  *rsa.PrivateKey

	FilePair
}
