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
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func WalletBalance(ctx context.Context, info *coins.TokenInfo, in *foxproxy.GetBalanceRequest) (*foxproxy.GetBalanceResponse, error) {
	err := tron.ValidAddress(info.Contract)
	if err != nil {
		return nil, fmt.Errorf("contract %v, %v, %v", info.Contract, tron.AddressInvalid, err)
	}

	bl := tron.EmptyTRC20
	if err := tron.ValidAddress(in.Address); err != nil {
		return nil, err
	}

	client := tron.Client()
	err = client.WithClient(info.LocalAPIs, info.PublicAPIs, func(c *tronclient.GrpcClient) (bool, error) {
		bl, err = c.TRC20ContractBalance(in.Address, info.Contract)
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
	balance, _ := f.Float64()

	return &foxproxy.GetBalanceResponse{
		Info: &foxproxy.BalanceInfo{
			Balance:    balance,
			BalanceStr: f.Text('f', tron.TRC20ACCURACY),
		},
	}, nil
}

func BuildTransaciton(ctx context.Context, info *coins.TokenInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
	err := tron.ValidAddress(tx.From)
	if err != nil {
		return nil, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}

	err = tron.ValidAddress(tx.To)
	if err != nil {
		return nil, fmt.Errorf("%v,%v", tron.AddressInvalid, err)
	}

	err = tron.ValidAddress(info.Contract)
	if err != nil {
		return nil, fmt.Errorf("contract %v, %v, %v", info.Contract, tron.AddressInvalid, err)
	}

	var txExtension *api.TransactionExtention
	client := tron.Client()
	err = client.WithClient(info.LocalAPIs, info.PublicAPIs, func(c *tronclient.GrpcClient) (bool, error) {
		txExtension, err = c.TRC20Send(
			tx.From,
			tx.To,
			info.Contract,
			tron.TRC20ToBigInt(tx.Amount),
			tron.TRC20FeeLimit,
		)
		return false, err
	})
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(txExtension)

	submitTx := coins.ToSubmitTx(tx)
	submitTx.Payload = payload

	return submitTx, nil
}
