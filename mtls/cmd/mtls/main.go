package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/bingoohuang/ngg/mtls"
	"github.com/spf13/cobra"
)

func main() {
	Execute()
}

var rootCmd = &cobra.Command{
	Use:   "mtls",
	Short: "mtls for mkcerts, demo server and demo client",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			clientCrt = filepath.Join(certsPath, "client.crt")
			clientKey = filepath.Join(certsPath, "client.key")
			serverCrt = filepath.Join(certsPath, "server.crt")
			serverKey = filepath.Join(certsPath, "server.key")
			rootCrt   = filepath.Join(certsPath, "ca.crt")
			rootKey   = filepath.Join(certsPath, "ca.key")
		)

		err := mtls.Verify("localhost", clientCrt, clientKey,
			serverCrt, serverKey, rootCrt, rootKey, true)
		if err != nil {
			return err
		}

		srv := HTTPServer(serverCrt, serverKey, rootCrt)
		defer srv.Close()

		SendRequest(srv.URL, clientCrt, clientKey, rootCrt)
		return nil
	},
}

func Execute() {
	p := rootCmd.PersistentFlags()
	p.StringVarP(&certsPath, "certs-path", "C", "./certs/", "router path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var certFile string

func init() {
	certInfoCmd.Flags().StringVarP(&certFile, "file", "f", "", "cert file")
	rootCmd.AddCommand(certInfoCmd)
}

var certInfoCmd = &cobra.Command{
	Use:   "certinfo",
	Short: "printing X509 TLS certificates",
	RunE:  func(*cobra.Command, []string) error { return mtls.CertInfo(certFile) },
}

var clientNames []string

func init() {
	mkCertsCmd.Flags().StringSliceVarP(&clientNames, "clients", "c", []string{"client"}, "client names")
	rootCmd.AddCommand(mkCertsCmd)
}

var mkCertsCmd = &cobra.Command{
	Use:   "mkcerts",
	Short: "Make certs",
	Long: `Make all certs
环境变量：
1. DNS_NAMES 指定域名, e.g. domain.demo
`,
	RunE: func(*cobra.Command, []string) error { return mtls.MakeCerts(certsPath, clientNames) },
}

var (
	port, sslPort         int
	serverPath, certsPath string
)

func init() {
	f := serverCmd.Flags()
	f.StringVarP(&serverPath, "server-path", "", "/", "router path")
	f.IntVarP(&port, "port", "p", 8080, "http listens port")
	f.IntVarP(&sslPort, "ssl-port", "", 8443, "https listens port")

	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Demo Server",
	Long:  `Demo Server`,
	RunE:  func(*cobra.Command, []string) error { return mtls.StartServer(serverPath, certsPath, port, sslPort) },
}

var (
	clientName string
	urlAddress string
)

func init() {
	f := clientCmd.Flags()
	f.StringVarP(&clientName, "client", "c", "Client1", "client name")
	f.StringVarP(&urlAddress, "url", "", "https://127.0.0.1:8443", "server https URL")

	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Demo Client",
	Long:  `Demo Client`,
	RunE:  func(*cobra.Command, []string) error { return mtls.StartClient(certsPath, clientName, urlAddress) },
}

// HTTPServer will start a mTLS enabled httptest server and return the test server.
// It requires server certificate's public and private key files and root certificate's public key file as arguments.
func HTTPServer(serverPublicKey, serverPrivateKey, rootPublicKey string) *httptest.Server {
	// server certificate.
	serverCert, err := tls.LoadX509KeyPair(serverPublicKey, serverPrivateKey)
	if err != nil {
		return nil
	}

	// root certificate.
	rootCert, err := os.ReadFile(rootPublicKey)
	if err != nil {
		log.Fatalf("failed to read root public key: %v", err)
	}
	rootCertPool := x509.NewCertPool()
	rootCertPool.AppendCertsFromPEM(rootCert)

	// httptest server with TLS config.
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mtls.PrintHeader(r)
		mtls.PrintTLSConnectionState(r.TLS)

		_, _ = fmt.Fprintln(w, "success!")
	}))
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    rootCertPool,
		ClientAuth:   mtls.If(mtls.GetEnvBool("CLIENT_AUTH_OFF"), tls.NoClientCert, tls.RequireAndVerifyClientCert),
	}
	server.StartTLS()
	return server
}

// SendRequest function will send a GET request to the server URL provided in the argument.
// It also requires client certificate's public and private key files and root certificate's public key file.
func SendRequest(serverURL, clientPublicKey, clientPrivateKey, rootPublicKey string) {
	// root certificate public key
	rootCert, errRead := os.ReadFile(rootPublicKey)
	if errRead != nil {
		log.Fatalf("failed to read public key: %v", errRead)
	}
	publicPemBlock, _ := pem.Decode(rootCert)
	rootPubCrt, errParse := x509.ParseCertificate(publicPemBlock.Bytes)
	if errParse != nil {
		log.Fatalf("failed to parse public key: %v", errParse)
	}

	rootCertpool := x509.NewCertPool()
	rootCertpool.AddCert(rootPubCrt)

	// client certificates.
	cert, err := tls.LoadX509KeyPair(clientPublicKey, clientPrivateKey)
	if err != nil {
		log.Fatalf("failed to load client certificate: %v", err)
	}

	// http client with root and client certificates.
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      rootCertpool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	r, err := client.Get(serverURL)
	if err != nil {
		log.Printf("failed to GET: %v", err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read body: %v", err)
		return
	}
	defer func(r io.ReadCloser) {
		_ = r.Close()
	}(r.Body)

	mtls.PrintTLSConnectionState(r.TLS)

	log.Printf("successful GET: %s", string(body))
}
