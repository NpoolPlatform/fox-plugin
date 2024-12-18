package declient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func (mgr *DEClientMGR) SendAndRecv(ctx context.Context, msgType foxproxy.MsgType, req interface{}, resp interface{}) (*foxproxy.StatusCode, error) {
	inPayload, err := json.Marshal(req)
	if err != nil {
		return foxproxy.StatusCode_StatusCodeMarshalErr.Enum(), err
	}

	recvChannel := make(chan MsgInfo)
	defer close(recvChannel)

	err = mgr.SendMsg(msgType, nil, &MsgInfo{Payload: inPayload}, &recvChannel)
	if err != nil {
		return foxproxy.StatusCode_StatusCodeFailed.Enum(), err
	}

	var recvMsg MsgInfo
	select {
	case <-ctx.Done():
		return foxproxy.StatusCode_StatusCodeFailed.Enum(), ctx.Err()
	case <-time.NewTimer(time.Second * 3).C:
		return foxproxy.StatusCode_StatusCodeFailed.Enum(), fmt.Errorf("timeout for recv response")
	case recvMsg = <-recvChannel:
	}

	if recvMsg.StatusCode.String() != foxproxy.StatusCode_StatusCodeSuccess.String() {
		if recvMsg.StatusMsg == nil {
			return recvMsg.StatusCode, fmt.Errorf("")
		}
		return recvMsg.StatusCode, fmt.Errorf(*recvMsg.StatusMsg)
	}

	err = json.Unmarshal(recvMsg.Payload, resp)
	if err != nil {
		return foxproxy.StatusCode_StatusCodeUnmarshalErr.Enum(), err
	}

	return foxproxy.StatusCode_StatusCodeSuccess.Enum(), nil
}
