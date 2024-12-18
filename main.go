package main

import (
	"context"
	"fmt"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/client"
	"github.com/NpoolPlatform/fox-plugin/pkg/config"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	"github.com/NpoolPlatform/go-service-framework/pkg/wlog"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

func main() {
	logger.Init(logger.DebugLevel, "./a.log")
	ctx, cancel := context.WithCancel(context.Background())
	go OnProxyConnection(ctx)
	TestSend()
	cancel()
}

func OnProxyConnection(ctx context.Context) {
	go func() {
		for i := 0; ; {
			select {
			case <-ctx.Done():
				fmt.Println("end", time.Second*(1<<i))
				return
			case <-time.NewTimer(time.Second * (1 << i)).C:
				fmt.Println("start", time.Second*(1<<i))
				if err := ConnectAndRecv(ctx); err != nil {
					logger.Sugar().Error(err)
					i++
				} else {
					i = 0
				}
				if i > 8 {
					i = 0
				}
			}
		}
	}()
	ctx.Done()
}

// will block on
func ConnectAndRecv(ctx context.Context) error {
	tlsConfig, err := client.LoadTLSConfig("/var/secure/client.a.crt", "/var/secure/client.a.key", "/var/secure/ca.crt")
	if err != nil {
		return wlog.Errorf("failed to get tls config, err: %v", err)
	}
	conn, err := client.GetGRPCConn(config.GetENV().Proxy, &tlsConfig)
	if err != nil {
		return wlog.Errorf("failed to get grpc connection, err: %v", err)

	}
	defer conn.Close()

	proxyClient := foxproxy.NewFoxProxyStreamClient(conn)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pcClient, err := proxyClient.DEStream(ctx)
	if err != nil {
		return wlog.Errorf("failed to get proxy connection, err: %v", err)
	}

	proxyConn, err := declient.RegisterDEClient(
		pcClient,
		foxproxy.ClientType_ClientTypePlugin,
		config.GetENV().Position,
		[]*foxproxy.CoinInfo{
			{
				Name:      "btc",
				CoinType:  foxproxy.CoinType_CoinTypealeo,
				ChainType: foxproxy.ChainType_Bitcoin,
			},
		})
	if err != nil {
		return wlog.Errorf("failed to get proxy connection, err: %v", err)
	}

	mgr := declient.GetDEClientMGR()
	mgr.AddDEClient(proxyConn)

	proxyConn.OnRecv()
	return nil
}

func TestSend() {
	mgr := declient.GetDEClientMGR()
	for i := 0; i < 3; i++ {
		time.Sleep(time.Second * 10)
		payload := []byte(fmt.Sprintf("send from plugin, idx: %v", i))
		recvChannel := make(chan declient.MsgInfo)
		err := mgr.SendMsg(
			foxproxy.MsgType_MsgTypeGetBalance,
			nil,
			&declient.MsgInfo{
				Payload: payload,
			},
			&recvChannel,
		)

		if err != nil {
			logger.Sugar().Error(err)
		}
		msg := <-recvChannel
		fmt.Println(string(msg.Payload))
	}
}
