package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/foxproxy"

	tronclient "github.com/Geapefurit/gotron-sdk/pkg/client"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron"
	"github.com/NpoolPlatform/fox-plugin/pkg/env"

	"github.com/Geapefurit/gotron-sdk/pkg/common"
	"github.com/Geapefurit/gotron-sdk/pkg/proto/api"
	"github.com/Geapefurit/gotron-sdk/pkg/proto/core"
	ct "github.com/NpoolPlatform/fox-plugin/pkg/types"
)

// redefine Code ,because github.com/Geapefurit/gotron-sdk/pkg/proto/core/Tron.pb.go line 564 spelling err
const (
	TransactionInfoSUCCESS = 0
	TransactionInfoFAILED  = 1
)

func WalletBalance(ctx context.Context, info *coins.TokenInfo, in *foxproxy.GetBalanceRequest) (*foxproxy.GetBalanceResponse, error) {
	if err := tron.ValidAddress(in.Address); err != nil {
		return nil, err
	}

	client := tron.Client()
	var bl int64
	err := client.WithClient(info.LocalAPIs, info.PublicAPIs, func(cli *tronclient.GrpcClient) (bool, error) {
		acc, err := cli.GetAccount(in.Address)
		if err != nil && strings.Contains(err.Error(), tron.AddressNotActive) {
			bl = tron.EmptyTRX
			return false, nil
		}
		if err != nil {
			return true, err
		}
		if acc == nil {
			return true, errors.New(tron.GetAccountFailed)
		}
		bl = acc.GetBalance()
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	f := tron.TRXToBigFloat(bl)
	blance, _ := f.Float64()

	return &foxproxy.GetBalanceResponse{
		Info: &foxproxy.BalanceInfo{
			Balance:    blance,
			BalanceStr: f.Text('f', tron.TRXACCURACY),
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

	from := tx.From
	to := tx.To
	amount := tron.TRXToInt(tx.Amount)

	client := tron.Client()

	var txExtension *api.TransactionExtention
	err = client.WithClient(info.LocalAPIs, info.PublicAPIs, func(cli *tronclient.GrpcClient) (bool, error) {
		_, err := cli.GetAccount(from)
		if err != nil {
			return true, err
		}
		if tron.TxFailErr(err) {
			return false, err
		}
		txExtension, err = cli.Transfer(from, to, amount)
		if err != nil {
			return true, err
		}
		if txExtension == nil {
			return false, errors.New(tron.BuildTransactionFailed)
		}
		return false, err
	})
	if err != nil {
		return nil, err
	}

	submitTx := coins.ToSubmitTx(tx)

	payload, err := json.Marshal(txExtension)
	if err != nil {
		return nil, err
	}

	submitTx.Payload = payload
	return submitTx, nil
}

func BroadcastTransaction(ctx context.Context, info *coins.TokenInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
	txExtension := &api.TransactionExtention{}

	err := json.Unmarshal(tx.Payload, txExtension)
	if err != nil {
		return nil, err
	}

	client := tron.Client()
	var result *api.Return
	err = client.WithClient(info.LocalAPIs, info.PublicAPIs, func(cli *tronclient.GrpcClient) (bool, error) {
		result, err = cli.Broadcast(txExtension.Transaction)
		if err != nil && result != nil && result.GetCode() == api.Return_TRANSACTION_EXPIRATION_ERROR {
			return false, err
		}
		if err != nil || result == nil {
			return true, err
		}
		return false, err
	})

	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("get result failed")
	}

	if api.Return_SUCCESS == result.Code && result.Result {
		submitTx := coins.ToSubmitTx(tx)
		txID := common.BytesToHexString(txExtension.GetTxid())
		bResp := &ct.BroadcastInfo{TxID: txID}
		payload, err := json.Marshal(bResp)
		if err != nil {
			return nil, err
		}

		submitTx.CID = &txID
		submitTx.Payload = payload
		return submitTx, nil
	}

	failCodes := []api.ReturnResponseCode{
		// api.Return_SUCCESS,
		api.Return_SIGERROR,
		api.Return_CONTRACT_VALIDATE_ERROR,
		api.Return_CONTRACT_EXE_ERROR,
		// api.Return_BANDWIDTH_ERROR=4,
		4,
		api.Return_DUP_TRANSACTION_ERROR,
		api.Return_TAPOS_ERROR,
		api.Return_TOO_BIG_TRANSACTION_ERROR,
		api.Return_TRANSACTION_EXPIRATION_ERROR,
		// api.Return_SERVER_BUSY,
		// api.Return_NO_CONNECTION,
		// api.Return_NOT_ENOUGH_EFFECTIVE_CONNECTION,
		api.Return_OTHER_ERROR,
	}
	for _, v := range failCodes {
		if v == result.Code {
			return nil, env.ErrTransactionFail
		}
	}

	return nil, errors.New(string(result.GetMessage()))
}

// done(on chain) => true
func SyncTxState(ctx context.Context, info *coins.TokenInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
	syncReq := &ct.BroadcastInfo{}
	err := json.Unmarshal(tx.Payload, syncReq)
	if err != nil {
		logger.Sugar().Errorw("SyncTxState", "Req", syncReq, "Error", err)
		return nil, err
	}
	client := tron.Client()

	var txInfo *core.TransactionInfo
	err = client.WithClient(info.LocalAPIs, info.PublicAPIs, func(cli *tronclient.GrpcClient) (bool, error) {
		txInfo, err = cli.GetTransactionInfoByID(syncReq.TxID)
		if err != nil {
			logger.Sugar().Errorw("SyncTxState", "Req", syncReq, "Error", err)
			return true, err
		}
		return false, err
	})

	submitTx := coins.ToSubmitTx(tx)
	submitTx.ExitCode = 1

	if txInfo == nil || err != nil {
		logger.Sugar().Errorw("SyncTxState", "Req", syncReq, "Info", txInfo, "Error", err, "Msg", "tx is syncing")
		return submitTx, nil
	}

	if txInfo.GetResult() != TransactionInfoSUCCESS {
		logger.Sugar().Errorw("SyncTxState", "Req", syncReq, "Info", txInfo, "Result", txInfo.GetResult())
		return nil, env.ErrTransactionFail
	}

	if txInfo.Receipt.GetResult() != core.Transaction_Result_SUCCESS && txInfo.Receipt.GetResult() != core.Transaction_Result_DEFAULT {
		logger.Sugar().Errorw("SyncTxState", "Req", syncReq, "Info", txInfo, "Result", txInfo.GetResult())
		return nil, env.ErrTransactionFail
	}

	submitTx.ExitCode = 0
	return submitTx, nil
}
