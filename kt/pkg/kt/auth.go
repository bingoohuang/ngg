package kt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/IBM/sarama"
)

type AuthConfig struct {
	AuthMode    string `json:"mode" default:"sasl" enum:"sasl,tls,tls-1way" help:"auth mode" persistent:"1"`
	Cert        string `json:"ca" help:"root ca cert file" persistent:"1"`
	ClientCert  string `json:"cert" help:"client cert file" persistent:"1"`
	ClientKey   string `json:"key" help:"client cert key" persistent:"1"`
	SASLUsr     string `json:"usr" help:"sasl user" persistent:"1"`
	SASLPwd     string `json:"pwd" help:"sasl password" persistent:"1"`
	SASLVersion *int   `json:"ver" help:"sasl KafkaVersion" default:"0" enum:"0,1" persistent:"1"`
}

func (t AuthConfig) SetupAuth(sc *sarama.Config) error {
	switch {
	case strings.EqualFold(t.AuthMode, "SASL") || t.SASLUsr != "":
		return t.setupSASL(sc)
	case strings.EqualFold(t.AuthMode, "TLS") || t.Cert != "":
		return t.setupAuthTLS(sc)
	case strings.EqualFold(t.AuthMode, "TLS-1way"):
		return t.setupAuthTLS1Way(sc)

	case t.AuthMode == "":
		return nil
	default:
		return fmt.Errorf("unsupport auth mode: %#v", t.AuthMode)
	}
}

func (t AuthConfig) setupSASL(sc *sarama.Config) error {
	sc.Net.SASL.Enable = true
	sc.Net.SASL.User = t.SASLUsr
	sc.Net.SASL.Password = t.SASLPwd
	sc.Net.SASL.Handshake = true
	sc.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	version, err := SASLVersion(sc.Version, t.SASLVersion)
	if err != nil {
		return err
	}
	sc.Net.SASL.Version = version
	return nil
}

func SASLVersion(kafkaVersion sarama.KafkaVersion, saslVersion *int) (int16, error) {
	if saslVersion == nil {
		if kafkaVersion.IsAtLeast(sarama.V1_0_0_0) {
			return sarama.SASLHandshakeV1, nil
		}
		return sarama.SASLHandshakeV0, nil
	}

	switch *saslVersion {
	case 0:
		return sarama.SASLHandshakeV0, nil
	case 1:
		return sarama.SASLHandshakeV1, nil
	default:
		return 0, errors.New("invalid SASL KafkaVersion")
	}
}

func (t AuthConfig) setupAuthTLS1Way(sc *sarama.Config) error {
	sc.Net.TLS.Enable = true
	sc.Net.TLS.Config = &tls.Config{}
	return nil
}

func (t AuthConfig) setupAuthTLS(sc *sarama.Config) error {
	tlsCfg, err := createTLSConfig(t.Cert, t.ClientCert, t.ClientKey)
	if err != nil {
		return err
	}

	sc.Net.TLS.Enable = true
	sc.Net.TLS.Config = tlsCfg

	return nil
}

func createTLSConfig(caCert, clientCert, certKey string) (*tls.Config, error) {
	if caCert == "" || clientCert == "" || certKey == "" {
		return nil, fmt.Errorf("a-cert, client-cert and client-key are required")
	}

	caString, err := os.ReadFile(caCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca-cert err=%v", err)
	}

	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caString); !ok {
		return nil, fmt.Errorf("unable to add ca-cert at %s to certificate pool", caCert)
	}

	cert, err := tls.LoadX509KeyPair(clientCert, certKey)
	if err != nil {
		return nil, err
	}

	tlsCfg := &tls.Config{RootCAs: caPool, Certificates: []tls.Certificate{cert}}
	return tlsCfg, nil
}
