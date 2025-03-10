package declient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient/types"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"github.com/google/uuid"
)

// nolint: revive
type DEClientMGR struct {
	recvChannel sync.Map
	connections []*DEClient
}

var cmgr *DEClientMGR

func GetDEClientMGR() *DEClientMGR {
	if cmgr == nil {
		cmgr = &DEClientMGR{
			recvChannel: sync.Map{},
		}
	}
	return cmgr
}

func (mgr *DEClientMGR) AddDEClient(conn *DEClient) {
	mgr.connections = append(mgr.connections, conn)
	conn.WatchRecv(mgr.DealDataElement)
	conn.WatchClose(mgr.deleteConnection)
}

func (mgr *DEClientMGR) GetClientInfos() []*foxproxy.ClientInfo {
	ret := []*foxproxy.ClientInfo{}
	for _, info := range mgr.connections {
		ret = append(ret, info.ClientInfo)
	}
	return ret
}

// delete conn from connectionMGR
func (mgr *DEClientMGR) deleteConnection(conn *DEClient) {
	for i := 0; i < len(mgr.connections); i++ {
		idx := len(mgr.connections) - 1 - i
		if mgr.connections[idx].ID == conn.ID {
			mgr.connections = append(mgr.connections[:idx], mgr.connections[idx+1:]...)
		}
	}
}

func (mgr *DEClientMGR) CloseAll() {
	for _, conn := range mgr.connections {
		conn.Close()
	}
}

// if recvChannel is not nil, recv response will send to it
// default value of statusCode is success
func (mgr *DEClientMGR) SendMsg(
	msgType foxproxy.MsgType,
	msgID *string,
	connID *string,
	msg *types.MsgInfo,
) error {
	var conn *DEClient
	conns := mgr.connections

	if len(conns) == 0 {
		return fmt.Errorf("cannot find any proxy connection")
	}

	if connID == nil {
		conn = conns[time.Now().Second()%len(conns)]
	} else {
		for _, _conn := range conns {
			if _conn.ID == *connID {
				conn = _conn
				break
			}
		}
		if conn == nil {
			return fmt.Errorf("cannot find any proxy connection,for %v", connID)
		}
	}
	return mgr.sendMsg(msgType, msgID, msg, conn)
}

// if recvChannel is not nil, recv response will send to it
// default value of statusCode is success
func (mgr *DEClientMGR) SendMsgWithConnID(
	msgType foxproxy.MsgType,
	connID string,
	msgID *string,
	msg *types.MsgInfo,
) error {
	var conn *DEClient
	for _, _conn := range mgr.connections {
		if _conn.ID == connID {
			conn = _conn
			break
		}
	}
	if conn == nil {
		return fmt.Errorf("cannot find any sider,for %v", connID)
	}

	return mgr.sendMsg(msgType, msgID, msg, conn)
}

// if recvChannel is not nil, recv response will send to it
// default value of statusCode is success
func (mgr *DEClientMGR) sendMsg(
	msgType foxproxy.MsgType,
	msgID *string,
	msg *types.MsgInfo,
	conn *DEClient,
) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if msg == nil {
		msg = &types.MsgInfo{}
	}

	if msgID == nil {
		_msgID := uuid.NewString()
		msgID = &_msgID
	}

	return conn.Send(&foxproxy.DataElement{
		ConnectID: conn.ID,
		MsgID:     *msgID,
		MsgType:   msgType,
		Payload:   msg.Payload,
		ErrMsg:    msg.ErrMsg,
	})
}

func (mgr *DEClientMGR) SendAndRecvRaw(ctx context.Context, msgType foxproxy.MsgType, inPayload []byte) ([]byte, error) {
	msgID := uuid.NewString()
	recvChannel := make(chan types.MsgInfo)
	mgr.recvChannel.Store(msgID, recvChannel)

	defer close(recvChannel)
	defer mgr.recvChannel.Delete(msgID)

	err := mgr.SendMsg(msgType, &msgID, nil, &types.MsgInfo{Payload: inPayload})
	if err != nil {
		return nil, wlog.WrapError(err)
	}

	var recvMsg types.MsgInfo
	select {
	case <-ctx.Done():
		return nil, wlog.WrapError(ctx.Err())
	case <-time.NewTimer(time.Second * 3).C:
		return nil, wlog.Errorf("timeout for recv response")
	case recvMsg = <-recvChannel:
	}

	if recvMsg.ErrMsg != nil && *recvMsg.ErrMsg != "" {
		return nil, wlog.Errorf(*recvMsg.ErrMsg)
	}
	return recvMsg.Payload, nil
}

func (mgr *DEClientMGR) SendAndRecv(ctx context.Context, msgType foxproxy.MsgType, req, resp interface{}) error {
	inPayload, err := json.Marshal(req)
	if err != nil {
		return wlog.WrapError(err)
	}

	outPayload, err := mgr.SendAndRecvRaw(ctx, msgType, inPayload)
	if err != nil {
		return wlog.WrapError(err)
	}
	if outPayload == nil || resp == nil {
		return nil
	}
	err = json.Unmarshal(outPayload, resp)
	if err != nil {
		return wlog.WrapError(err)
	}

	return nil
}

func (mgr *DEClientMGR) DealDataElement(data *foxproxy.DataElement) {
	if ch, ok := mgr.recvChannel.LoadAndDelete(data.MsgID); ok {
		select {
		case <-time.NewTimer(time.Second).C:
		case ch.(chan types.MsgInfo) <- types.MsgInfo{
			Payload: data.Payload,
			ErrMsg:  data.ErrMsg,
		}:
		}
	}

	if data.MsgType == foxproxy.MsgType_MsgTypeResponse {
		return
	}

	var resp *types.MsgInfo
	h, err := handler.GetTokenMGR().GetDEHandler(data.MsgType)
	if err != nil {
		logger.Sugar().Error(err)
		statusMsg := err.Error()
		resp = &types.MsgInfo{
			ErrMsg: &statusMsg,
		}
	}

	if h != nil {
		resp = h(context.Background(), data)
	}
	if resp == nil {
		return
	}

	err = mgr.SendMsgWithConnID(foxproxy.MsgType_MsgTypeResponse, data.ConnectID, &data.MsgID, resp)
	if err != nil {
		logger.Sugar().Error(err)
		return
	}
}
