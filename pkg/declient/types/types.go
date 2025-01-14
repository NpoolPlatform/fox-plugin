package types

import "github.com/NpoolPlatform/message/npool/foxproxy"

type MsgInfo struct {
	Payload  []byte
	ErrMsg   *string
	CoinInfo *foxproxy.CoinInfo
}
