package handler

import (
	"context"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient/types"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

type DEHandlerFunc func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo
type TxHandlerFunc func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo

type TokenMGR struct {
	deHandlers    map[foxproxy.MsgType]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc
	txHandlers    map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]TxHandlerFunc
	tokenInfos    map[string]*coins.TokenInfo    // from code register
	depTokenInfos map[string]*coins.DepTokenInfo // from deployer
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
		deHandlers:    make(map[foxproxy.MsgType]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc),
		txHandlers:    make(map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]TxHandlerFunc),
		tokenInfos:    make(map[string]*coins.TokenInfo),
		depTokenInfos: make(map[string]*coins.DepTokenInfo),
	}
}
