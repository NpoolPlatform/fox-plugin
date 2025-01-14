package declient

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"github.com/google/uuid"
)

type DEClient struct {
	foxproxy.FoxProxyStream_DEStreamClient
	*foxproxy.ClientInfo
	ctx          context.Context
	cancel       context.CancelFunc
	onCloseFuncs []func(conn *DEClient)
	recvHandlers []func(data *foxproxy.DataElement)
	closeOnce    sync.Once
}

func RegisterDEClient(
	stream foxproxy.FoxProxyStream_DEStreamClient,
	clientType foxproxy.ClientType,
	position string,
	infos []*foxproxy.CoinInfo,
) (*DEClient, error) {
	select {
	case <-time.NewTicker(time.Second * 3).C:
		return nil, wlog.Errorf("timeout for register connection")
	default:
		if len(infos) == 0 {
			return nil, wlog.Errorf("have no infos")
		}

		msgID := uuid.NewString()
		connID := uuid.NewString()
		connInfo := &foxproxy.ClientInfo{
			ClientType: clientType,
			ID:         connID,
			Infos:      infos,
			Position:   position,
		}
		payload, err := json.Marshal(connInfo)
		if err != nil {
			return nil, wlog.WrapError(err)
		}

		err = stream.Send(&foxproxy.DataElement{
			ConnectID:  connID,
			MsgID:      msgID,
			Payload:    payload,
			StatusCode: foxproxy.StatusCode_StatusCodeSuccess,
		})
		if err != nil {
			return nil, wlog.WrapError(err)
		}

		data, err := stream.Recv()
		if err != nil {
			return nil, wlog.WrapError(err)
		}

		if data.StatusCode != foxproxy.StatusCode_StatusCodeSuccess {
			return nil, wlog.Errorf("failed to register to proxy, err: %v", data.StatusMsg)
		}

		ctx, cancel := context.WithCancel(stream.Context())
		return &DEClient{
			ClientInfo:                    connInfo,
			FoxProxyStream_DEStreamClient: stream,
			ctx:                           ctx,
			cancel:                        cancel,
		}, nil
	}
}

func (conn *DEClient) WatchClose(onClose func(conn *DEClient)) {
	conn.onCloseFuncs = append(conn.onCloseFuncs, onClose)
}

func (conn *DEClient) WatchRecv(onRecv func(data *foxproxy.DataElement)) {
	conn.recvHandlers = append(conn.recvHandlers, onRecv)
}

func (conn *DEClient) OnRecv() {
	defer conn.cancel()
	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
			data, err := conn.Recv()
			if err != nil {
				logger.Sugar().Error(err)
				return
			}
			for _, recvHandler := range conn.recvHandlers {
				go recvHandler(data)
			}
		}
	}
}

func (conn *DEClient) Close() {
	conn.closeOnce.Do(func() {
		for _, onClose := range conn.onCloseFuncs {
			onClose(conn)
		}
		if err := conn.FoxProxyStream_DEStreamClient.CloseSend(); err != nil {
			logger.Sugar().Warn(err)
		}

		conn.cancel()
		_, ok := <-conn.ctx.Done()
		logger.Sugar().Warn(ok)

		logger.Sugar().Warnf(
			"connection is closed, clientType: %v, ID: %v, Position: %v",
			conn.ClientType,
			conn.Position,
			conn.Position)
	})
}
