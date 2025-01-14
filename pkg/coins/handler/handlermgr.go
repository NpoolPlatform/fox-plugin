package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient/types"
	"github.com/NpoolPlatform/fox-plugin/pkg/utils"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"gopkg.in/yaml.v2"
)

type DEHandlerFunc func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo

type TokenMGR struct {
	msgHandlers   map[foxproxy.MsgType]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc
	txHandlers    map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc
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
		msgHandlers:   make(map[foxproxy.MsgType]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc),
		txHandlers:    make(map[foxproxy.TransactionState]map[foxproxy.ChainType]map[foxproxy.CoinType]DEHandlerFunc),
		tokenInfos:    make(map[string]*coins.TokenInfo),
		depTokenInfos: make(map[string]*coins.DepTokenInfo),
	}
}

// duplicate register is not allowed
func (mgr *TokenMGR) RegisterTokenInfo(info *coins.TokenInfo) error {
	if _, ok := mgr.tokenInfos[info.Name]; ok {
		return fmt.Errorf("already exist token: %v", info.Name)
	}
	mgr.tokenInfos[info.Name] = info
	return nil
}

// duplicate register is not allowed
func (mgr *TokenMGR) RegisterDepTokenInfo(infos []*coins.DepTokenInfo) error {
	for _, info := range infos {
		tempInfo := mgr.GetTokenInfo(info.TempName)
		if tempInfo == nil {
			return fmt.Errorf("invalid temp name: %v", info.TempName)
		}

		if _, ok := mgr.depTokenInfos[info.Name]; ok {
			return fmt.Errorf("already exist token: %v", info.Name)
		}

		if info.ENV == coins.CoinNetMain {
			_tempInfo := *tempInfo
			_tempInfo.TaskInterval = info.TaskInterval
			_tempInfo.LocalAPIs = info.LocalAPIs
			_tempInfo.PublicAPIs = info.PublicAPIs
			info.TokenInfo = _tempInfo
		} else if info.Name == info.TempName {
			return fmt.Errorf("this name(%v) cannot be used for non-mainnet", info.Name)
		}

		mgr.depTokenInfos[info.Name] = info
	}
	return nil
}

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

func (mgr *TokenMGR) GetTokenInfo(name string) *coins.TokenInfo {
	return mgr.tokenInfos[name]
}

func (mgr *TokenMGR) GetTokenInfos() []*coins.TokenInfo {
	ret := []*coins.TokenInfo{}
	for _, v := range mgr.tokenInfos {
		ret = append(ret, v)
	}
	return ret
}

func (mgr *TokenMGR) GetCoinInfos() []*foxproxy.CoinInfo {
	ret := []*foxproxy.CoinInfo{}
	for _, v := range mgr.tokenInfos {
		ret = append(ret, &foxproxy.CoinInfo{
			Name:      v.Name,
			TempName:  v.Name,
			CoinType:  v.CoinType,
			ChainType: v.ChainType,
			ENV:       v.ENV,
		})
	}
	return ret
}

func (mgr *TokenMGR) GetDepTokenInfo(name string) *coins.DepTokenInfo {
	return mgr.depTokenInfos[name]
}

func (mgr *TokenMGR) GetDepTokenInfos() []*coins.DepTokenInfo {
	ret := []*coins.DepTokenInfo{}
	for _, v := range mgr.depTokenInfos {
		ret = append(ret, v)
	}
	return ret
}

func (mgr *TokenMGR) GetDepCoinInfos() []*foxproxy.CoinInfo {
	ret := []*foxproxy.CoinInfo{}
	for _, v := range mgr.depTokenInfos {
		ret = append(ret, &foxproxy.CoinInfo{
			Name:      v.Name,
			TempName:  v.TempName,
			CoinType:  v.CoinType,
			ChainType: v.ChainType,
			ENV:       v.ENV,
		})
	}
	return ret
}

func (mgr *TokenMGR) GetTokenRegisterCoinInfos() []*foxproxy.RegisterCoinInfo {
	ret := []*foxproxy.RegisterCoinInfo{}
	for _, v := range mgr.depTokenInfos {
		ret = append(ret, &foxproxy.RegisterCoinInfo{
			Name:                v.Name,
			Unit:                v.Unit,
			ENV:                 v.ENV,
			ChainType:           v.ChainType,
			ChainNativeUnit:     v.ChainNativeUnit,
			ChainAtomicUnit:     v.ChainAtomicUnit,
			ChainUnitExp:        v.ChainUnitExp,
			GasType:             v.GasType,
			ChainID:             v.ChainID,
			ChainNickname:       v.ChainNickname,
			ChainNativeCoinName: v.ChainNativeCoinName,
		})
	}
	return ret
}

func (mgr *TokenMGR) GetDEHandler(msgType foxproxy.MsgType) (DEHandlerFunc, error) {
	if msgType != foxproxy.MsgType_MsgTypeDealTxs {
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

	return func(ctx context.Context, data *foxproxy.DataElement) *types.MsgInfo { return nil }, nil
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

type DepTokenInfos struct {
	Infos []utils.Entry `yaml:"TokenInfos"`
}

func (mgr *TokenMGR) RegisterDepTokenInfosFromYaml(yamlFile string) error {
	yamlBytes, err := os.ReadFile(yamlFile)
	if err != nil {
		return wlog.WrapError(err)
	}

	depTokenEntrys := DepTokenInfos{}
	err = yaml.Unmarshal(yamlBytes, &depTokenEntrys)
	if err != nil {
		return wlog.WrapError(err)
	}

	modifiableFileds := coins.GetModifiableFileds()

	depTokenInfos := []*coins.DepTokenInfo{}
	for _, envEntry := range depTokenEntrys.Infos {
		if _, ok := envEntry["TempName"]; !ok {
			return wlog.Errorf("have no TempName field")
		}
		tempName, ok := envEntry["TempName"].(string)
		if !ok {
			return wlog.Errorf("value of TempName is not string")
		}

		depEntry := make(utils.Entry)
		for _, filedName := range modifiableFileds {
			if v, ok := envEntry[filedName]; ok {
				depEntry[filedName] = v
			}
		}

		if info := mgr.GetTokenInfo(tempName); info != nil {
			infoMap, err := utils.StructToMap(coins.DepTokenInfo{TempName: info.Name, TokenInfo: *info})
			if err != nil {
				return wlog.WrapError(err)
			}

			ret := &coins.DepTokenInfo{}
			err = utils.MapToStruct(utils.CoverEntry(infoMap, depEntry), ret)
			if err != nil {
				return wlog.WrapError(err)
			}
			depTokenInfos = append(depTokenInfos, ret)
		}
	}

	return mgr.RegisterDepTokenInfo(depTokenInfos)
}
