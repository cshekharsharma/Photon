package system

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// LoadTLSCredentials loads TLS credentials from cert, key, and CA files.
func LoadTLSCredentials(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key: %w", err)
	}

	caCert, err := os.ReadFile(caFile) // #nosec G304 -- caller intentionally supplies the CA bundle path.
	if err != nil {
		return nil, fmt.Errorf("failed to read CA cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append CA certs")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}, nil
}
