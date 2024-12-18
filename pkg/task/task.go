package task

import (
	"context"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/declient"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	v1 "github.com/NpoolPlatform/message/npool/basetypes/v1"
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
			in := &foxproxy.RegisterCoinInfo{
				Name:                "TestCoin",
				Unit:                "TC",
				ENV:                 "test",
				ChainType:           foxproxy.ChainType_Aleo,
				ChainNativeUnit:     "TC",
				ChainAtomicUnit:     "tTC",
				ChainUnitExp:        6,
				GasType:             v1.GasType_FixedGas,
				ChainID:             "ROCK",
				ChainNickname:       "TestCoin",
				ChainNativeCoinName: "TestCoin",
			}
			out := &foxproxy.EmptyResponse{}
			statusCode, err := mgr.SendAndRecv(context.Background(), foxproxy.MsgType_MsgTypeRegisterCoin, in, out)
			if err != nil {
				logger.Sugar().Error(statusCode, err)
				continue
			}
			return
		}
	}
}

func Run(ctx context.Context) {
	go declient.GetDEClientMGR().StartDEStream(ctx)
	go declient.GetDEClientMGR().StartDEStream(ctx)
	go RegisterCoin(ctx)
}
