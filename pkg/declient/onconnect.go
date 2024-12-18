package declient

import (
	"context"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/client"
	"github.com/NpoolPlatform/fox-plugin/pkg/config"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func (mgr *DEClientMGR) StartDEStream(ctx context.Context) {
	go func() {
		for i := 0; ; {
			select {
			case <-ctx.Done():
				return
			case <-time.NewTimer(time.Second * (1 << i)).C:
				if err := mgr.connectAndRecv(ctx); err != nil {
					logger.Sugar().Error(err)
					i++
				} else {
					i = 0
				}
				if i > 8 {
					i = 0
				}
			}
		}
	}()
	ctx.Done()
}

// will block on
func (mgr *DEClientMGR) connectAndRecv(ctx context.Context) error {
	tlsConfig, err := client.LoadTLSConfig("/var/certs/client.a.crt", "/var/certs/client.a.key", "/var/certs/ca.crt")
	if err != nil {
		return wlog.Errorf("failed to get tls config, err: %v", err)
	}
	conn, err := client.GetGRPCConn(config.GetENV().Proxy, &tlsConfig)
	if err != nil {
		return wlog.Errorf("failed to get grpc connection, err: %v", err)

	}
	defer conn.Close()

	proxyClient := foxproxy.NewFoxProxyStreamClient(conn)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pcClient, err := proxyClient.DEStream(ctx)
	if err != nil {
		return wlog.Errorf("failed to get proxy connection, err: %v", err)
	}

	proxyConn, err := RegisterDEClient(
		pcClient,
		foxproxy.ClientType_ClientTypePlugin,
		config.GetENV().Position,
		[]*foxproxy.CoinInfo{
			{
				Name:      "btc",
				CoinType:  foxproxy.CoinType_CoinTypealeo,
				ChainType: foxproxy.ChainType_Bitcoin,
			},
		})
	if err != nil {
		return wlog.Errorf("failed to get proxy connection, err: %v", err)
	}

	mgr.AddDEClient(proxyConn)

	proxyConn.OnRecv()
	return nil
}
