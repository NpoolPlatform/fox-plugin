package handler

import (
	"context"
	"fmt"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func (mgr *TokenMGR) RegisterPluginTxHandler(
	state foxproxy.TransactionState,
	info *coins.TokenInfo,
	handler func(ctx context.Context, info *coins.TokenInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error),
) {
	txHandler := func(ctx context.Context, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
		info := mgr.GetDepTokenInfo(tx.Name)
		return handler(ctx, &info.TokenInfo, tx)
	}

	if _, ok := mgr.pluginTxHandlers[state]; !ok {
		mgr.pluginTxHandlers[state] = make(map[foxproxy.ChainType]map[foxproxy.CoinType]PluginTxHandlerFunc)
	}
	if _, ok := mgr.pluginTxHandlers[state][info.ChainType]; !ok {
		mgr.pluginTxHandlers[state][info.ChainType] = make(map[foxproxy.CoinType]PluginTxHandlerFunc)
	}
	mgr.pluginTxHandlers[state][info.ChainType][info.CoinType] = txHandler
}

func (mgr *TokenMGR) RegisterSignTxHandler(
	state foxproxy.TransactionState,
	info *coins.TokenInfo,
	handler func(ctx context.Context, info *foxproxy.CoinInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error),
) {
	txHandler := func(ctx context.Context, info *foxproxy.CoinInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
		return handler(ctx, info, tx)
	}

	if _, ok := mgr.signTxHandlers[state]; !ok {
		mgr.signTxHandlers[state] = make(map[foxproxy.ChainType]map[foxproxy.CoinType]SignTxHandlerFunc)
	}
	if _, ok := mgr.signTxHandlers[state][info.ChainType]; !ok {
		mgr.signTxHandlers[state][info.ChainType] = make(map[foxproxy.CoinType]SignTxHandlerFunc)
	}
	mgr.signTxHandlers[state][info.ChainType][info.CoinType] = txHandler
}

func (mgr *TokenMGR) GetPluginTxHandler(state foxproxy.TransactionState, chainType foxproxy.ChainType, coinType foxproxy.CoinType) (PluginTxHandlerFunc, error) {
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

func (mgr *TokenMGR) GetSignTxHandler(state foxproxy.TransactionState, chainType foxproxy.ChainType, coinType foxproxy.CoinType) (SignTxHandlerFunc, error) {
	_, ok := mgr.signTxHandlers[state]
	if !ok {
		return nil, fmt.Errorf("have no handler for tx state: %v", state)
	}
	_, ok = mgr.signTxHandlers[state][chainType]
	if !ok {
		return nil, fmt.Errorf("have no handler for tx state: %v - chaintype: %v", state, chainType)
	}
	h, ok := mgr.signTxHandlers[state][chainType][coinType]
	if !ok {
		return nil, fmt.Errorf("have no handler for tx state: %v - chaintype: %v, cointype: %v", state, chainType, coinType)
	}
	return h, nil
}
