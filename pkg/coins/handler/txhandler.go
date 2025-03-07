package handler

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func (mgr *TokenMGR) RegisterTxHandler(
	state foxproxy.TransactionState,
	info *coins.TokenInfo,
	handler func(ctx context.Context, info *coins.TokenInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error),
) {
	txHandler := func(ctx context.Context, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
		_info := info
		depInfo := mgr.GetDepTokenInfo(tx.Name)
		if depInfo != nil {
			_info = &depInfo.TokenInfo
		}
		return handler(ctx, _info, tx)
	}

	if _, ok := mgr.pluginTxHandlers[state]; !ok {
		mgr.pluginTxHandlers[state] = make(map[foxproxy.ChainType]map[foxproxy.CoinType]PluginTxHandlerFunc)
	}
	if _, ok := mgr.pluginTxHandlers[state][info.ChainType]; !ok {
		mgr.pluginTxHandlers[state][info.ChainType] = make(map[foxproxy.CoinType]PluginTxHandlerFunc)
	}
	mgr.pluginTxHandlers[state][info.ChainType][info.CoinType] = txHandler
}

func (mgr *TokenMGR) GetTxHandler(state foxproxy.TransactionState, chainType foxproxy.ChainType, coinType foxproxy.CoinType) (PluginTxHandlerFunc, error) {
	_, ok := mgr.pluginTxHandlers[state]
	if !ok {
		return nil, fmt.Errorf("have no handler for tx state: %v", state)
	}
	_, ok = mgr.pluginTxHandlers[state][chainType]
	if !ok {
		return nil, fmt.Errorf("have no handler for tx state: %v - chaintype: %v", state, chainType)
	}
	h, ok := mgr.pluginTxHandlers[state][chainType][coinType]
	if !ok {
		return nil, fmt.Errorf("have no handler for tx state: %v - chaintype: %v, cointype: %v", state, chainType, coinType)
	}
	return h, nil
}
