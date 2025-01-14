package task

import (
	"context"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/client"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/fox-plugin/pkg/config"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

// keeping register coin util successul
func RegisterCoin(ctx context.Context) {
	mgr := declient.GetDEClientMGR()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.NewTimer(time.Second * 3).C:
			in := handler.GetTokenMGR().GetTokenRegisterCoinInfos()
			out := &foxproxy.EmptyResponse{}
			err := mgr.SendAndRecv(context.Background(), foxproxy.MsgType_MsgTypeRegisterCoin, in, out)
			if err != nil {
				logger.Sugar().Error(err)
				continue
			}
			return
		}
	}
}

func Run(ctx context.Context) {
	tlsConfig, err := client.LoadTLSConfig("/var/certs/client.a.crt", "/var/certs/client.a.key", "/var/certs/ca.crt")
	if err != nil {
		panic(wlog.Errorf("failed to get tls config, err: %v", err))
	}
	go declient.GetDEClientMGR().StartDEStream(
		ctx,
		foxproxy.ClientType_ClientTypePlugin,
		config.GetENV().Proxy,
		config.GetENV().Position,
		tlsConfig)
	go declient.GetDEClientMGR().StartDEStream(ctx,
		foxproxy.ClientType_ClientTypePlugin,
		config.GetENV().Proxy,
		config.GetENV().Position,
		tlsConfig)
	go RegisterCoin(ctx)
}
