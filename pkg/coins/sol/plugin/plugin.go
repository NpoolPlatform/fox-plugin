package plugin

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/utils"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func WalletBalance(ctx context.Context, tokenInfo *coins.TokenInfo, in *foxproxy.GetBalanceRequest) (*foxproxy.GetBalanceResponse, error) {
	logger.Sugar().Info(utils.PrettyStruct(tokenInfo))

	return &foxproxy.GetBalanceResponse{
		Info: &foxproxy.BalanceInfo{
			Balance:    2.333,
			BalanceStr: "Happy New Year!",
		},
	}, nil
}
