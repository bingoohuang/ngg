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
)

// StartServer start a demo server for http and https.
func StartServer(serverPath, certsPath string, port, sslPort int) error {
	// Set up a /hello resource handler
	handler := http.NewServeMux()
	handler.HandleFunc(serverPath, helloHandler)

	// Listen to port 8080 and wait
	go func() {
		server := http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler,
		}
		fmt.Printf("(HTTP) Listen on :%d\n", port)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("(HTTP) error listening to port: %v", err)
		}
	}()

	// load CA certificate file and add it to list of client CAs
	caCertFile, err := os.ReadFile(filepath.Join(certsPath, "ca.crt"))
	if err != nil {
		return fmt.Errorf("reading CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertFile)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs:        caCertPool,
		ClientAuth:       If(GetEnvBool("CLIENT_AUTH_OFF"), tls.NoClientCert, tls.RequireAndVerifyClientCert),
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		InsecureSkipVerify: GetEnvBool("INSECURE_SKIP_VERIFY"),
	}

	// serve on port 8443 of local host
	server := http.Server{
		Addr:      fmt.Sprintf(":%d", sslPort),
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("(HTTPS) Listen on :%d\n", sslPort)
	if err := server.ListenAndServeTLS(
		filepath.Join(certsPath, "server.crt"),
		filepath.Join(certsPath, "server.key")); err != nil {
		return fmt.Errorf(" listening to port: %w", err)
	}

	return nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	PrintHeader(r)
	PrintTLSConnectionState(r.TLS)

	log.Print(">>>>>>>>>>>>>>>>> End <<<<<<<<<<<<<<<<<<")
	fmt.Println("")
	// Write "Hello, world!" to the response body
	_, _ = io.WriteString(w, "Hello, world!\n")
}

func PrintHeader(r *http.Request) {
	log.Print(">>>>>>>>>>>>>>>> Header <<<<<<<<<<<<<<<<")
	// Loop over header names
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			log.Printf("%v:%v", name, value)
		}
	}
}

func PrintTLSConnectionState(state *tls.ConnectionState) {
	if state == nil {
		return
	}

	log.Print(">>>>>>>>>>>>>>>> tls.ConnectionState <<<<<<<<<<<<<<<<")

	log.Printf("Version: %x", state.Version)
	log.Printf("HandshakeComplete: %t", state.HandshakeComplete)
	log.Printf("DidResume: %t", state.DidResume)
	log.Printf("CipherSuite: %x", state.CipherSuite)
	log.Printf("NegotiatedProtocol: %s", state.NegotiatedProtocol)

	log.Print("Certificate chain:")
	for i, cert := range state.PeerCertificates {
		s := cert.Subject
		u := cert.Issuer
		log.Printf(" %d subject: /C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s", i,
			s.Country, s.Province, s.Locality, s.Organization, s.OrganizationalUnit, s.CommonName)
		log.Printf("   issuer : /C=%v/ST=%v/L=%v/O=%v/OU=%v/CN=%s",
			u.Country, u.Province, u.Locality, u.Organization, u.OrganizationalUnit, u.CommonName)
	}
}
