package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/declient"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const grpcPort = 50023

var serverStream foxproxy.FoxProxyStream_DEStreamServer

func (s *Server) DEStream(stream foxproxy.FoxProxyStream_DEStreamServer) error {
	RegisterDEServer(stream)
	serverStream = stream
	<-serverStream.Context().Done()
	return nil
}

type Server struct {
	foxproxy.UnimplementedFoxProxyStreamServer
}

// for test
func MockOnServer(ctx context.Context, grpcPort int) {
	fmt.Println("start to mock server")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	defer server.Stop()

	foxproxy.RegisterFoxProxyStreamServer(server, &Server{})

	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	<-ctx.Done()
	fmt.Println("end to mock server")
}

// for test
func RegisterDEServer(stream foxproxy.FoxProxyStream_DEStreamServer) error {
	select {
	case <-time.NewTicker(time.Second * 3).C:
		return wlog.Errorf("timeout for register connection")
	default:
		data, err := stream.Recv()
		if err != nil {
			return wlog.WrapError(err)
		}

		statusCode := foxproxy.StatusCode_StatusCodeSuccess
		statusMsg := ""

		connInfo := &foxproxy.ClientInfo{}
		err = json.Unmarshal(data.Payload, connInfo)
		if err != nil {
			statusCode = foxproxy.StatusCode_StatusCodeFailed
			statusMsg = err.Error()
		}

		if statusCode == foxproxy.StatusCode_StatusCodeSuccess && data.ConnectID != connInfo.ID {
			statusCode = foxproxy.StatusCode_StatusCodeFailed
			statusMsg = err.Error()
		}

		err = stream.Send(&foxproxy.DataElement{
			ConnectID:  data.ConnectID,
			MsgID:      data.MsgID,
			StatusCode: statusCode,
			StatusMsg:  &statusMsg,
		})
		if err != nil {
			return wlog.WrapError(err)
		}

		if statusCode != foxproxy.StatusCode_StatusCodeSuccess {
			return wlog.Errorf(statusMsg)
		}
		return nil
	}
}

func TestDEClientMGR(t *testing.T) {
	logger.Init(logger.DebugLevel, "./a.log")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go MockOnServer(ctx, grpcPort)

	// for server start
	time.Sleep(time.Second)
	mgr := declient.GetDEClientMGR()
	mgr.StartDEStream(ctx, fmt.Sprintf("localhost:%d", grpcPort), "test", nil)
	// for client connect
	time.Sleep(time.Second * 2)

	infos := mgr.GetClientInfos()
	if !assert.NotEqual(t, 0, len(infos)) {
		return
	}

	payload := []byte("sssssss")
	err := mgr.SendMsg(foxproxy.MsgType_MsgTypeDefault, nil, &declient.MsgInfo{
		Payload: payload,
	}, nil)
	if !assert.Nil(t, err) {
		return
	}

	dataEle, err := serverStream.Recv()
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, payload, dataEle.Payload)

	req := foxproxy.RegisterCoinInfo{Name: "sdsfasdfa"}
	resp := foxproxy.RegisterCoinInfo{}
	go func() {
		dataEle, err := serverStream.Recv()
		if !assert.Nil(t, err) {
			os.Exit(1)
		}
		err = serverStream.Send(dataEle)
		if !assert.Nil(t, err) {
			os.Exit(1)
		}
	}()
	statucCode, err := mgr.SendAndRecv(ctx, foxproxy.MsgType_MsgTypeGetBalance, &req, &resp)
	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, statucCode.String(), foxproxy.StatusCode_StatusCodeSuccess.String())
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, req.ChainID, resp.ChainID)

	mgr.CloseAll()
	err = mgr.SendMsg(foxproxy.MsgType_MsgTypeDefault, nil, &declient.MsgInfo{
		Payload: payload,
	}, nil)
	assert.NotNil(t, err)
}