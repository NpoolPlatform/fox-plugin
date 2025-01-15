package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient/types"
	"github.com/NpoolPlatform/fox-plugin/pkg/utils"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func (mgr *TokenMGR) RegisterPluginHandler(
	msgType foxproxy.MsgType,
	info *coins.TokenInfo,
	in interface{},
	handler func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error),
) {
	deHandler := func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo {
		info := mgr.GetDepTokenInfo(data.CoinInfo.Name)
		if info != nil {
			if info.CoinType != data.CoinInfo.CoinType || info.ChainType != data.CoinInfo.ChainType {
				statusMsg := fmt.Sprintf(
					"cannot match cointype or chaintype: name: %v, cointype: %v-%v, chaintype: %v-%v",
					info.Name,
					info.CoinType,
					data.CoinInfo.CoinType,
					info.ChainType,
					data.CoinInfo.ChainType)

				return &types.MsgInfo{
					Payload:  nil,
					ErrMsg:   &statusMsg,
					CoinInfo: data.CoinInfo,
				}
			}
		}

		// decode payload to requeast
		// run handler
		// and encode result to payload
		outPayload, err := func(data *foxproxy.DataElement, info *coins.TokenInfo, in interface{}) ([]byte, error) {
			inData := utils.Copy(in)
			err := json.Unmarshal(data.Payload, inData)
			if err != nil {
				return nil, err
			}

			out, err := handler(ctx, data.CoinInfo, info, inData)
			if err != nil {
				return nil, err
			}

			outPayload, err := json.Marshal(out)
			if err != nil {
				return nil, err
			}
			return outPayload, nil
		}(data, &info.TokenInfo, in)

		statusMsg := ""
		if err != nil {
			statusMsg = err.Error()
		}

		return &types.MsgInfo{
			Payload:  outPayload,
			ErrMsg:   &statusMsg,
			CoinInfo: data.CoinInfo,
		}
	}

	if _, ok := mgr.msgHandlers[msgType]; !ok {
		mgr.msgHandlers[msgType] = make(map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc)
	}
	if _, ok := mgr.msgHandlers[msgType][info.ChainType]; !ok {
		mgr.msgHandlers[msgType][info.ChainType] = make(map[foxproxy.CoinType]DEHandlerFunc)
	}
	mgr.msgHandlers[msgType][info.ChainType][info.CoinType] = deHandler
}

func (mgr *TokenMGR) RegisterSignHandler(
	msgType foxproxy.MsgType,
	info *coins.TokenInfo,
	in interface{},
	handler func(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req interface{}) (interface{}, error),
) {
	deHandler := func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo {
		info := mgr.GetTokenInfo(data.CoinInfo.TempName)
		if info != nil {
			if info.CoinType != data.CoinInfo.CoinType || info.ChainType != data.CoinInfo.ChainType {
				statusMsg := fmt.Sprintf(
					"cannot match cointype or chaintype: tempname: %v, cointype: %v-%v, chaintype: %v-%v",
					data.CoinInfo.TempName,
					info.CoinType,
					data.CoinInfo.CoinType,
					info.ChainType,
					data.CoinInfo.ChainType)

				return &types.MsgInfo{
					Payload:  nil,
					ErrMsg:   &statusMsg,
					CoinInfo: data.CoinInfo,
				}
			}
		}

		// decode payload to requeast
		// run handler
		// and encode result to payload
		outPayload, err := func(data *foxproxy.DataElement, info *coins.TokenInfo, in interface{}) ([]byte, error) {
			inData := utils.Copy(in)
			err := json.Unmarshal(data.Payload, inData)
			if err != nil {
				return nil, err
			}

			out, err := handler(ctx, data.CoinInfo, info, inData)
			if err != nil {
				return nil, err
			}

			outPayload, err := json.Marshal(out)
			if err != nil {
				return nil, err
			}
			return outPayload, nil
		}(data, info, in)

		statusMsg := ""
		if err != nil {
			statusMsg = err.Error()
		}

		return &types.MsgInfo{
			Payload:  outPayload,
			ErrMsg:   &statusMsg,
			CoinInfo: data.CoinInfo,
		}
	}

	if _, ok := mgr.msgHandlers[msgType]; !ok {
		mgr.msgHandlers[msgType] = make(map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc)
	}
	if _, ok := mgr.msgHandlers[msgType][info.ChainType]; !ok {
		mgr.msgHandlers[msgType][info.ChainType] = make(map[foxproxy.CoinType]DEHandlerFunc)
	}
	mgr.msgHandlers[msgType][info.ChainType][info.CoinType] = deHandler
}

func (mgr *TokenMGR) GetDEHandler(msgType foxproxy.MsgType) (DEHandlerFunc, error) {
	return func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo {
		if data.CoinInfo == nil {
			statusMsg := "the required coin info was not given."
			return &types.MsgInfo{
				Payload: nil,
				ErrMsg:  &statusMsg,
			}
		}

		h, err := mgr.GetMsgHandler(msgType, data.CoinInfo.ChainType, data.CoinInfo.CoinType)
		if err != nil {
			statusMsg := err.Error()
			return &types.MsgInfo{
				Payload: nil,
				ErrMsg:  &statusMsg,
			}
		}
		return h(ctx, data)
	}, nil
}

func (mgr *TokenMGR) GetMsgHandler(msgType foxproxy.MsgType, chainType foxproxy.ChainType, coinType foxproxy.CoinType) (DEHandlerFunc, error) {
	_, ok := mgr.msgHandlers[msgType]
	if !ok {
		return nil, fmt.Errorf("have no handler for msgtype: %v", msgType)
	}
	_, ok = mgr.msgHandlers[msgType][chainType]
	if !ok {
		return nil, fmt.Errorf("have no handler for msgtype: %v - chaintype: %v", msgType, chainType)
	}
	h, ok := mgr.msgHandlers[msgType][chainType][coinType]
	if !ok {
		return nil, fmt.Errorf("have no handler for msgtype: %v - chaintype: %v, cointype: %v", msgType, chainType, coinType)
	}
	return h, nil
}
