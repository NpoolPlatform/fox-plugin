package sign

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/utils"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

// createAccount ..
func CreateAccount(ctx context.Context, coinInfo *foxproxy.CoinInfo, in *foxproxy.CreateWalletRequest) (out *foxproxy.CreateWalletResponse, err error) {
	logger.Sugar().Error(utils.PrettyStruct(coinInfo))
	return &foxproxy.CreateWalletResponse{Info: &foxproxy.WalletInfo{Address: "Good Day!"}}, nil
}
