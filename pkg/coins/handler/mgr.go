package handler

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient/types"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

type DEHandlerFunc func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo
type PluginTxHandlerFunc func(ctx context.Context, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error)
type SignTxHandlerFunc func(ctx context.Context, info *foxproxy.CoinInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error)

type TokenMGR struct {
	deHandlers       map[foxproxy.MsgType]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc
	pluginTxHandlers map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]PluginTxHandlerFunc
	signTxHandlers   map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]SignTxHandlerFunc
	tokenInfos       map[string]*coins.TokenInfo    // from code register
	depTokenInfos    map[string]*coins.DepTokenInfo // from deployer
}

var hmgr *TokenMGR

func GetTokenMGR() *TokenMGR {
	if hmgr == nil {
		hmgr = newTokenMGR()
	}
	return hmgr
}

func newTokenMGR() *TokenMGR {
	return &TokenMGR{
		deHandlers:       make(map[foxproxy.MsgType]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc),
		pluginTxHandlers: make(map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]PluginTxHandlerFunc),
		signTxHandlers:   make(map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]SignTxHandlerFunc),
		tokenInfos:       make(map[string]*coins.TokenInfo),
		depTokenInfos:    make(map[string]*coins.DepTokenInfo),
	}
}
