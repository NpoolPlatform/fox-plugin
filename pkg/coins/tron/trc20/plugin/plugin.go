package trc20

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tronclient "github.com/Geapefurit/gotron-sdk/pkg/client"
	"github.com/Geapefurit/gotron-sdk/pkg/proto/api"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron"
	ct "github.com/NpoolPlatform/fox-plugin/pkg/types"
)

func WalletBalance(ctx context.Context, in []byte, info *coins.TokenInfo) (out []byte, err error) {
	wbReq := &ct.WalletBalanceRequest{}
	err = json.Unmarshal(in, wbReq)
	if err != nil {
		return nil, err
	}

	contract := info.Contract
	err = tron.ValidAddress(contract)
	if err != nil {
		return nil, fmt.Errorf("contract %v, %v, %v", contract, tron.AddressInvalid, err)
	}

	bl := tron.EmptyTRC20
	if err := tron.ValidAddress(wbReq.Address); err != nil {
		return nil, err
	}

	client := tron.Client()
	err = client.WithClient(info.LocalAPIs, info.PublicAPIs, func(c *tronclient.GrpcClient) (bool, error) {
		bl, err = c.TRC20ContractBalance(wbReq.Address, contract)
		if err != nil && strings.Contains(err.Error(), tron.AddressNotActive) {
			bl = tron.EmptyTRC20
			return false, nil
		}
		if err != nil {
			return true, err
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	f := tron.TRC20ToBigFloat(bl)
	wbResp := &ct.WalletBalanceResponse{}

	wbResp.Balance, _ = f.Float64()
	wbResp.BalanceStr = f.Text('f', tron.TRC20ACCURACY)

	out, err = json.Marshal(wbResp)

	return out, err
}

func BuildTransaciton(ctx context.Context, in []byte, info *coins.TokenInfo) (out []byte, err error) {
	baseInfo := &ct.BaseInfo{}
	err = json.Unmarshal(in, baseInfo)
	if err != nil {
		return nil, err
	}

	err = tron.ValidAddress(baseInfo.From)
	if err != nil {
		return nil, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}

	err = tron.ValidAddress(baseInfo.To)
	if err != nil {
		return nil, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}

	contract := info.Contract
	err = tron.ValidAddress(contract)
	if err != nil {
		return nil, fmt.Errorf("contract %v, %v, %v", contract, tron.AddressInvalid, err)
	}

	var txExtension *api.TransactionExtention
	client := tron.Client()
	err = client.WithClient(info.LocalAPIs, info.PublicAPIs, func(c *tronclient.GrpcClient) (bool, error) {
		txExtension, err = c.TRC20Send(
			baseInfo.From,
			baseInfo.To,
			contract,
			tron.TRC20ToBigInt(baseInfo.Value),
			tron.TRC20FeeLimit,
		)
		return false, err
	})
	if err != nil {
		return nil, err
	}
	signTx := &tron.SignMsgTx{
		Base:        *baseInfo,
		TxExtension: txExtension,
	}

	return json.Marshal(signTx)
}
