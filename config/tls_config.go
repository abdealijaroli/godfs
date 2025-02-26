package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
)

func DevTransport() net.Listener {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		panic(err)
	}
	return listener
}

func LoadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load keypair: %v", err)
	}

	caData, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read CA: %v", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caData) {
		return nil, fmt.Errorf("parse CA: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
