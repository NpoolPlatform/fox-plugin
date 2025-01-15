package register

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/sol"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/sol/plugin"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/sol/sign"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func init() {
	mgr := handler.GetTokenMGR()
	if err := mgr.RegisterTokenInfo(sol.SolanaToken); err != nil {
		panic(err)
	}
	mgr.RegisterPluginDEHandler(
		foxproxy.MsgType_MsgTypeGetBalance,
		sol.SolanaToken,
		&foxproxy.GetBalanceRequest{},
		func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error) {
			return plugin.WalletBalance(ctx, info, req.(*foxproxy.GetBalanceRequest))
		})
	mgr.RegisterSignDEHandler(
		foxproxy.MsgType_MsgTypeCreateWallet,
		sol.SolanaToken,
		&foxproxy.CreateWalletRequest{},
		func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error) {
			return sign.CreateAccount(ctx, coinInfo, req.(*foxproxy.CreateWalletRequest))
		})
}
