package task

import (
	"context"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func PluginPullTXs(ctx context.Context, txChan chan *foxproxy.Transaction) {
	clientMGR := declient.GetDEClientMGR()
	tokenMGR := handler.GetTokenMGR()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.NewTimer(time.Second * 3).C:
			coinInfos := tokenMGR.GetDepCoinInfos()
			txs := &[]*foxproxy.Transaction{}
			err := clientMGR.SendAndRecv(ctx, foxproxy.MsgType_MsgTypeAssginPluginTxs, coinInfos, txs)
			if err != nil {
				logger.Sugar().Error(err)
				continue
			}
			for _, tx := range *txs {
				txChan <- tx
			}
		}
	}
}

func PluginDealTxWorker(ctx context.Context, txChan chan *foxproxy.Transaction) {
	tokenMgr := handler.GetTokenMGR()
	declientMgr := declient.GetDEClientMGR()
	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-txChan:
			info := tokenMgr.GetDepTokenInfo(tx.Name)
			submitTx := func() *foxproxy.SubmitTransaction {
				_submitTx := &foxproxy.SubmitTransaction{
					TransactionID: tx.TransactionID,
					Payload:       tx.Payload,
					CID:           &tx.CID,
					State:         tx.State,
					LockTime:      tx.LockTime,
					ExitCode:      -1,
				}
				handler, err := tokenMgr.GetTxHandler(tx.State, info.ChainType, info.CoinType)
				if err != nil {
					logger.Sugar().Error(err)
					return _submitTx
				}
				submitTx, err := handler(ctx, tx)
				if err != nil {
					logger.Sugar().Error(err)
					return _submitTx
				}
				return submitTx
			}()

			err := declientMgr.SendAndRecv(ctx, foxproxy.MsgType_MsgTypeSubmitTx, submitTx, nil)
			if err != nil {
				logger.Sugar().Error(err)
			}
		}
	}
}

func SignPullTXs(ctx context.Context, txChan chan *foxproxy.Transaction) {
	clientMGR := declient.GetDEClientMGR()
	tokenMGR := handler.GetTokenMGR()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.NewTimer(time.Second * 3).C:
			coinInfos := tokenMGR.GetCoinInfos()
			txs := &[]*foxproxy.Transaction{}
			err := clientMGR.SendAndRecv(ctx, foxproxy.MsgType_MsgTypeAssginSignTxs, coinInfos, txs)
			if err != nil {
				logger.Sugar().Error(err)
				continue
			}
			for _, tx := range *txs {
				txChan <- tx
			}
		}
	}
}

func SignDealTxWorker(ctx context.Context, txChan chan *foxproxy.Transaction) {
	tokenMgr := handler.GetTokenMGR()
	declientMgr := declient.GetDEClientMGR()
	for {
		select {
		case <-ctx.Done():
			return
		case tx := <-txChan:
			info := tokenMgr.GetTokenInfo(tx.Name)
			submitTx := func() *foxproxy.SubmitTransaction {
				_submitTx := &foxproxy.SubmitTransaction{
					TransactionID: tx.TransactionID,
					Payload:       tx.Payload,
					CID:           &tx.CID,
					State:         tx.State,
					LockTime:      tx.LockTime,
					ExitCode:      -1,
				}
				handler, err := tokenMgr.GetTxHandler(tx.State, info.ChainType, info.CoinType)
				if err != nil {
					logger.Sugar().Error(err)
					return _submitTx
				}
				submitTx, err := handler(ctx, tx)
				if err != nil {
					logger.Sugar().Error(err)
					return _submitTx
				}
				return submitTx
			}()

			err := declientMgr.SendAndRecv(ctx, foxproxy.MsgType_MsgTypeSubmitTx, submitTx, nil)
			if err != nil {
				logger.Sugar().Error(err)
			}
		}
	}
}
