package tron

import (
	"github.com/Geapefurit/gotron-sdk/pkg/proto/api"
	ct "github.com/NpoolPlatform/sphinx-plugin/pkg/types"
)

type SignMsgTx struct {
	Base        ct.BaseInfo               `json:"base"`
	TxExtension *api.TransactionExtention `json:"tx_extension"`
}

type BroadcastRequest struct {
	TxExtension *api.TransactionExtention `json:"tx_extension"`
}
