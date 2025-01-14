package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
	rlk   sync.RWMutex
	conns sync.Map
)

// GetGRPCConn get grpc client conn
func GetGRPCConn(conn string, tlsCfg credentials.TransportCredentials) (*grpc.ClientConn, error) {
	if conn == "" {
		return nil, fmt.Errorf("conn is empty")
	}

	rlk.Lock()
	_conn, err := checkRemove(conn)
	rlk.Unlock()
	if err == nil {
		return _conn, nil
	}

	// log warn

	return getConn(conn, tlsCfg)
}

func getConn(target string, tlsCfg credentials.TransportCredentials) (*grpc.ClientConn, error) {
	rlk.Lock()
	defer rlk.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	secureOpt := grpc.WithInsecure()
	if tlsCfg != nil {
		secureOpt = grpc.WithTransportCredentials(tlsCfg)
	}

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithBlock(),
		secureOpt,
	)
	if err != nil {
		return nil, err
	}

	connState := conn.GetState()
	if connState != connectivity.Idle && connState != connectivity.Ready {
		return nil, fmt.Errorf("get conn state not ready: %v", connState)
	}

	conns.Store(target, conn)
	return conn, nil
}

// out set ctx
func checkRemove(target string) (conn *grpc.ClientConn, err error) {
	_conn, ok := conns.Load(target)
	if !ok {
		return nil, fmt.Errorf("server: %v not build conn", target)
	}

	healthOK := false
	defer func() {
		if !healthOK {
			_conn.(*grpc.ClientConn).Close()
			conns.Delete(target)
		}
	}()

	cli := grpc_health_v1.NewHealthClient(_conn.(*grpc.ClientConn))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	hth, err := cli.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return nil, err
	}

	if hth.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return nil, fmt.Errorf("server: %v is down", target)
	}

	healthOK = true
	return _conn.(*grpc.ClientConn), nil
}

func LoadTLSConfig(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certification: %w", err)
	}

	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("faild to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), nil
}
