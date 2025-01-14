package declient

import (
	"context"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/client"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"google.golang.org/grpc/credentials"
)

func (mgr *DEClientMGR) StartDEStream(ctx context.Context, clientType foxproxy.ClientType, proxyHost, position string, tlsCfg credentials.TransportCredentials) {
	go func() {
		for i := 0; ; {
			select {
			case <-ctx.Done():
				return
			case <-time.NewTimer(time.Second * (1 << i)).C:
				if err := mgr.connectAndRecv(ctx, clientType, proxyHost, position, tlsCfg); err != nil {
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
func (mgr *DEClientMGR) connectAndRecv(ctx context.Context, clientType foxproxy.ClientType, proxyHost, position string, tlsCfg credentials.TransportCredentials) error {
	conn, err := client.GetGRPCConn(proxyHost, tlsCfg)
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

	var coinInfo []*foxproxy.CoinInfo
	if clientType == foxproxy.ClientType_ClientTypePlugin {
		coinInfo = handler.GetTokenMGR().GetDepCoinInfos()
	} else {
		coinInfo = handler.GetTokenMGR().GetCoinInfos()
	}

	proxyConn, err := RegisterDEClient(
		pcClient,
		clientType,
		position,
		coinInfo,
	)
	if err != nil {
		return wlog.Errorf("failed to get proxy connection, err: %v", err)
	}

	mgr.AddDEClient(proxyConn)

	proxyConn.OnRecv()
	return nil
}
