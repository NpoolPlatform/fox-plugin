package register

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/plugin"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/sign"
	trc20plugin "github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/trc20/plugin"
	trc20sign "github.com/NpoolPlatform/fox-plugin/pkg/coins/tron/trc20/sign"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func init() {
	mgr := handler.GetTokenMGR()
	// ################### TRON ####################
	if err := mgr.RegisterTokenInfo(tron.TronToken); err != nil {
		panic(err)
	}

	mgr.RegisterDEHandler(
		foxproxy.MsgType_MsgTypeGetBalance,
		tron.TronToken,
		&foxproxy.GetBalanceRequest{},
		func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error) {
			return plugin.WalletBalance(ctx, info, req.(*foxproxy.GetBalanceRequest))
		})

	mgr.RegisterDEHandler(
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

	// ################### TRON USDT ####################
	if err := mgr.RegisterTokenInfo(tron.USDTToken); err != nil {
		panic(err)
	}

	mgr.RegisterDEHandler(
		foxproxy.MsgType_MsgTypeGetBalance,
		tron.USDTToken,
		&foxproxy.GetBalanceRequest{},
		func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error) {
			return trc20plugin.WalletBalance(ctx, info, req.(*foxproxy.GetBalanceRequest))
		})

	mgr.RegisterDEHandler(
		foxproxy.MsgType_MsgTypeCreateWallet,
		tron.USDTToken,
		&foxproxy.CreateWalletRequest{},
		func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error) {
			return trc20sign.CreateTrc20Account(ctx, coinInfo, info, req.(*foxproxy.CreateWalletRequest))
		})

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStatePrepare,
		tron.USDTToken,
		trc20plugin.BuildTransaciton,
	)

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStateSign,
		tron.USDTToken,
		trc20sign.SignTrc20MSG,
	)

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStateBroadcast,
		tron.USDTToken,
		plugin.BroadcastTransaction,
	)

	mgr.RegisterTxHandler(
		foxproxy.TransactionState_TransactionStateSync,
		tron.USDTToken,
		plugin.SyncTxState,
	)
}
