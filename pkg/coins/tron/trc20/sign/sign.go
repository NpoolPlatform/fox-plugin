package trc20

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	tron "github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/sign"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func SignTrc20MSG(ctx context.Context, info *coins.TokenInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
	return tron.SignTronMSG(ctx, info, tx)
}

func CreateTrc20Account(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req *foxproxy.CreateWalletRequest) (*foxproxy.CreateWalletResponse, error) {
	return tron.CreateTronAccount(ctx, info.S3KeyPrxfix, req)
}
