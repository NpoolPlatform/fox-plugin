package declient

import (
	"fmt"
	"sync"
	"time"

	"github.com/NpoolPlatform/message/npool/foxproxy"
	"github.com/google/uuid"
)

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

type MsgInfo struct {
	Payload    []byte
	StatusCode *foxproxy.StatusCode
	StatusMsg  *string
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
	connID *string,
	msg *MsgInfo,
	recvChannel *chan MsgInfo,
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
	return mgr.sendMsg(msgType, nil, msg, conn, recvChannel)
}

// if recvChannel is not nil, recv response will send to it
// default value of statusCode is success
func (mgr *DEClientMGR) sendMsg(
	msgType foxproxy.MsgType,
	msgID *string,
	msg *MsgInfo,
	conn *DEClient,
	recvChannel *chan MsgInfo,
) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if msg == nil {
		msg = &MsgInfo{}
	}

	if msgID == nil {
		_msgID := uuid.NewString()
		msgID = &_msgID
	}

	if recvChannel != nil {
		mgr.recvChannel.Store(*msgID, *recvChannel)
	}

	if msg.StatusCode == nil {
		msg.StatusCode = foxproxy.StatusCode_StatusCodeSuccess.Enum()
	}

	return conn.Send(&foxproxy.DataElement{
		ConnectID:  conn.ID,
		MsgID:      *msgID,
		MsgType:    msgType,
		Payload:    msg.Payload,
		StatusCode: *msg.StatusCode,
		StatusMsg:  msg.StatusMsg,
	})
}

func (mgr *DEClientMGR) DealDataElement(data *foxproxy.DataElement) {
	if ch, ok := mgr.recvChannel.LoadAndDelete(data.MsgID); ok {
		select {
		case <-time.NewTimer(time.Second).C:
		case ch.(chan MsgInfo) <- MsgInfo{
			Payload:    data.Payload,
			StatusCode: &data.StatusCode,
			StatusMsg:  data.StatusMsg,
		}:
		}
	}

	// handler, err := GetDEHandlerMGR().GetDEHandler(data.MsgType)
	// if err != nil {
	// 	logger.Sugar().Error(err)
	// 	return
	// }

	// err = handler(data)
	// if err != nil {
	// 	logger.Sugar().Error(err)
	// }
}