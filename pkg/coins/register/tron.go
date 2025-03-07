package register

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/plugin"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/sign"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func init() {
	mgr := handler.GetTokenMGR()
	if err := mgr.RegisterTokenInfo(tron.TronToken); err != nil {
		panic(err)
	}

	mgr.RegisterPluginDEHandler(
		foxproxy.MsgType_MsgTypeGetBalance,
		tron.TronToken,
		&foxproxy.GetBalanceRequest{},
		func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error) {
			return plugin.WalletBalance(ctx, info, req.(*foxproxy.GetBalanceRequest))
		})

	mgr.RegisterSignDEHandler(
		foxproxy.MsgType_MsgTypeCreateWallet,
		tron.TronToken,
		&foxproxy.CreateWalletRequest{},
		func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error) {
			return sign.CreateTrxAccount(ctx, coinInfo, info, req.(*foxproxy.CreateWalletRequest))
		})

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStatePrepare,
		tron.TronToken,
		plugin.BuildTransaciton,
	)

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStateSign,
		tron.TronToken,
		sign.SignTronMSG,
	)

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStateBroadcast,
		tron.TronToken,
		plugin.BroadcastTransaction,
	)

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStateSync,
		tron.TronToken,
		plugin.SyncTxState,
	)
}
