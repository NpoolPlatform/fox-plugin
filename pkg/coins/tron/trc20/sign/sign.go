package trc20

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	tron "github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/sign"
)

const s3KeyPrxfix = "usdttrc20/"

func SignTrc20MSG(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	return tron.SignTronMSG(ctx, s3KeyPrxfix, in)
}

func CreateTrc20Account(ctx context.Context, in []byte, tokenInfo *coins.TokenInfo) (out []byte, err error) {
	// return tron.CreateTronAccount(ctx, tokenInfo.S3KeyPrxfix, nil)
	return
}
