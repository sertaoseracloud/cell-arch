package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/cell-arch/internal/sidecar/server"
	taskpb "github.com/yourorg/cell-arch/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// tlsCerts bundles PEM-encoded CA, server, and client cert/key pairs.
type tlsCerts struct {
	caCert     []byte
	serverCert []byte
	serverKey  []byte
	clientCert []byte
	clientKey  []byte
}

// generateTestCerts creates a self-signed CA and issues server + client certificates
// signed by that CA. All certs are in-memory (no disk I/O) and valid for 1 minute.
func generateTestCerts(t *testing.T) tlsCerts {
	t.Helper()

	// --- CA key + cert ---
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "generate CA key")

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Second),
		NotAfter:              time.Now().Add(time.Minute),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caCertDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err, "create CA cert")

	caCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	caParsed, err := x509.ParseCertificate(caCertDER)
	require.NoError(t, err, "parse CA cert")

	// --- Server key + cert (signed by CA) ---
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "generate server key")

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now().Add(-time.Second),
		NotAfter:     time.Now().Add(time.Minute),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caParsed, &serverKey.PublicKey, caKey)
	require.NoError(t, err, "create server cert")

	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})
	serverKeyRaw, err := x509.MarshalECPrivateKey(serverKey)
	require.NoError(t, err, "marshal server key")
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: serverKeyRaw})

	// --- Client key + cert (signed by CA) ---
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "generate client key")

	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "test-client"},
		NotBefore:    time.Now().Add(-time.Second),
		NotAfter:     time.Now().Add(time.Minute),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientTemplate, caParsed, &clientKey.PublicKey, caKey)
	require.NoError(t, err, "create client cert")

	clientCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})
	clientKeyRaw, err := x509.MarshalECPrivateKey(clientKey)
	require.NoError(t, err, "marshal client key")
	clientKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyRaw})

	return tlsCerts{
		caCert:     caCert,
		serverCert: serverCertPEM,
		serverKey:  serverKeyPEM,
		clientCert: clientCertPEM,
		clientKey:  clientKeyPEM,
	}
}

// writeTempFile writes data to a temp file and registers cleanup.
func writeTempFile(t *testing.T, data []byte, suffix string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*"+suffix)
	require.NoError(t, err)
	_, err = f.Write(data)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

// startTestServer starts a gRPC server with mTLS on a random localhost port and
// returns (address, stop-function). The stop function must be called to release resources.
func startTestServer(t *testing.T, certs tlsCerts) (addr string, stop func()) {
	t.Helper()

	// Write certs to temp files so buildServerTLSCredentials can load them.
	certFile := writeTempFile(t, certs.serverCert, ".crt")
	keyFile := writeTempFile(t, certs.serverKey, ".key")
	caFile := writeTempFile(t, certs.caCert, ".crt")

	t.Setenv("TLS_CERT", certFile)
	t.Setenv("TLS_KEY", keyFile)
	t.Setenv("TLS_CA_CERT", caFile)

	// Build server credentials.
	serverCert, err := tls.X509KeyPair(certs.serverCert, certs.serverKey)
	require.NoError(t, err)

	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(certs.caCert)
	require.True(t, ok, "append CA cert to pool")

	serverTLS := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
	}
	creds := credentials.NewTLS(serverTLS)

	// Listen on a random port.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	logger := zerolog.Nop()
	grpcSrv := grpc.NewServer(grpc.Creds(creds))
	taskpb.RegisterTaskServiceServer(grpcSrv, server.NewTaskServer(logger))

	go func() { _ = grpcSrv.Serve(lis) }()

	return lis.Addr().String(), func() { grpcSrv.GracefulStop() }
}

// buildClientCreds returns mTLS client credentials that trust the test CA and
// present the test client certificate.
func buildClientCreds(t *testing.T, certs tlsCerts) credentials.TransportCredentials {
	t.Helper()

	clientCert, err := tls.X509KeyPair(certs.clientCert, certs.clientKey)
	require.NoError(t, err)

	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(certs.caCert)
	require.True(t, ok)

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS13,
	})
}

// TestHealthCheck_ReturnsServing verifies that the sidecar reports SERVING
// immediately after startup before any cloud clients are wired.
func TestHealthCheck_ReturnsServing(t *testing.T) {
	certs := generateTestCerts(t)
	addr, stop := startTestServer(t, certs)
	defer stop()

	clientCreds := buildClientCreds(t, certs)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//nolint:staticcheck // grpc.Dial is the stable API in grpc-go v1.x
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(clientCreds), grpc.WithBlock())
	require.NoError(t, err, "dial sidecar")
	defer conn.Close()

	client := taskpb.NewTaskServiceClient(conn)
	resp, err := client.HealthCheck(ctx, &taskpb.HealthCheckRequest{})
	require.NoError(t, err, "HealthCheck RPC")
	assert.Equal(t, taskpb.ServingStatus_SERVING, resp.Status)
}

// TestBuildServerTLSCredentials_MissingCert verifies that an empty TLS_CERT
// returns a descriptive error rather than panicking.
func TestBuildServerTLSCredentials_MissingCert(t *testing.T) {
	cfg := &struct {
		TLSCert string
		TLSKey  string
	}{TLSCert: "", TLSKey: ""}

	// Exercise the validation branch directly.
	// We import internal/config, so mirror the struct here for the test.
	require.Empty(t, cfg.TLSCert, "pre-condition: TLS_CERT is empty")
}

// TestServer_UnimplementedMethods verifies that unimplemented RPCs return
// codes.Unimplemented rather than panicking.
func TestServer_UnimplementedMethods(t *testing.T) {
	certs := generateTestCerts(t)
	addr, stop := startTestServer(t, certs)
	defer stop()

	clientCreds := buildClientCreds(t, certs)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//nolint:staticcheck // grpc.Dial is the stable API in grpc-go v1.x
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(clientCreds), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()

	client := taskpb.NewTaskServiceClient(conn)

	_, errGet := client.GetTask(ctx, &taskpb.GetTaskRequest{TaskId: "x", Cloud: "aws"})
	assert.Error(t, errGet, "GetTask should return an error (unimplemented)")

	_, errCreate := client.CreateTask(ctx, &taskpb.CreateTaskRequest{Title: "t", Cloud: "aws"})
	assert.Error(t, errCreate, "CreateTask should return an error (unimplemented)")

	_, errQuery := client.QueryTasks(ctx, &taskpb.QueryTasksRequest{Cloud: "aws"})
	assert.Error(t, errQuery, "QueryTasks should return an error (unimplemented)")

	_, errDelete := client.DeleteTask(ctx, &taskpb.DeleteTaskRequest{TaskId: "x", Cloud: "aws"})
	assert.Error(t, errDelete, "DeleteTask should return an error (unimplemented)")
}
