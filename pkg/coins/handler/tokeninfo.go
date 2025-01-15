package handler

import (
	"fmt"
	"os"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/utils"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"gopkg.in/yaml.v2"
)

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
