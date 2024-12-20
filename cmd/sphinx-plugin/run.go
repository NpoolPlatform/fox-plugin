package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/config"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient"
	"github.com/NpoolPlatform/fox-plugin/pkg/task"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var runCmd = &cli.Command{
	Name:    "run",
	Aliases: []string{"r"},
	Usage:   "Run fox Plugin daemon",
	After: func(c *cli.Context) error {
		return logger.Sync()
	},
	Before: func(ctx *cli.Context) error {
		// TODO: elegent set or get env
		config.SetENV(&config.ENVInfo{
			LocalWalletAddr:  localWalletAddr,
			PublicWalletAddr: publicWalletAddr,
			Proxy:            proxyAddress,
			SyncInterval:     syncInterval,
			Contract:         contract,
			LogDir:           logDir,
			LogLevel:         logLevel,
			WanIP:            wanIP,
			Position:         position,
			BuildChainServer: buildChainServer,
		})
		err := logger.Init(
			logger.DebugLevel,
			filepath.Join(config.GetENV().LogDir, "fox-plugin.log"),
			zap.AddCallerSkip(1),
		)
		if err != nil {
			panic(fmt.Errorf("fail to init logger: %v", err))
		}
		return nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "local-wallet",
			Aliases:     []string{"lw"},
			Usage:       "wallet address of local",
			EnvVars:     []string{"ENV_COIN_LOCAL_API"},
			Required:    true,
			Value:       "",
			Destination: &localWalletAddr,
		},
		&cli.StringFlag{
			Name:        "public-wallet",
			Aliases:     []string{"pw"},
			Usage:       "wallet address of public",
			EnvVars:     []string{"ENV_COIN_PUBLIC_API"},
			Required:    true,
			Value:       "",
			Destination: &publicWalletAddr,
		},
		// proxy address
		&cli.StringFlag{
			Name:        "proxy",
			Aliases:     []string{"p"},
			Usage:       "address of fox proxy",
			EnvVars:     []string{"ENV_PROXY"},
			Required:    true,
			Value:       "",
			Destination: &proxyAddress,
		},
		// sync interval
		&cli.Int64Flag{
			Name:        "sync-interval",
			Aliases:     []string{"si"},
			Usage:       "interval seconds of sync transaction on chain status",
			EnvVars:     []string{"ENV_SYNC_INTERVAL"},
			Value:       0,
			Destination: &syncInterval,
		},
		// contract id
		&cli.StringFlag{
			Name:        "contract",
			Aliases:     []string{"c"},
			Usage:       "id of contract",
			EnvVars:     []string{"ENV_CONTRACT"},
			Value:       "",
			Destination: &contract,
		},
		// log level
		&cli.StringFlag{
			Name:        "level",
			Aliases:     []string{"L"},
			Usage:       "level support debug|info|warning|error",
			EnvVars:     []string{"ENV_LOG_LEVEL"},
			Value:       "debug",
			DefaultText: "debug",
			Destination: &logLevel,
		},
		// log path
		&cli.StringFlag{
			Name:        "log",
			Aliases:     []string{"l"},
			Usage:       "log dir",
			EnvVars:     []string{"ENV_LOG_DIR"},
			Value:       "/var/log",
			DefaultText: "/var/log",
			Destination: &logDir,
		},
		// wan ip
		&cli.StringFlag{
			Name:        "wan-ip",
			Aliases:     []string{"w"},
			Usage:       "wan ip",
			EnvVars:     []string{"ENV_WAN_IP"},
			Required:    true,
			Value:       "",
			Destination: &wanIP,
		},
		// position
		&cli.StringFlag{
			Name:        "position",
			Aliases:     []string{"po"},
			Usage:       "position",
			EnvVars:     []string{"ENV_POSITION"},
			Required:    true,
			Value:       "",
			Destination: &position,
		},
		// position
		&cli.StringFlag{
			Name:        "build-chain-server",
			Aliases:     []string{"b"},
			Usage:       "build-chain server address",
			EnvVars:     []string{"ENV_BUILD_CHAIN_SERVER"},
			Required:    false,
			Value:       "",
			Destination: &buildChainServer,
		},
	},
	Action: func(c *cli.Context) error {
		logger.Sugar().Infof(
			"run plugin wanIP: %v, Position %v",
			config.GetENV().WanIP,
			config.GetENV().Position,
		)

		sig := make(chan os.Signal, 1)
		signal.Notify(
			sig,
			syscall.SIGABRT,
			syscall.SIGBUS,
			syscall.SIGFPE,
			syscall.SIGILL,
			syscall.SIGINT,
			syscall.SIGQUIT,
			syscall.SIGSEGV,
			syscall.SIGTERM,
		)

		ctx, cancel := context.WithCancel(c.Context)
		defer cancel()
		defer declient.GetDEClientMGR().CloseAll()

		go task.Run(ctx)

		<-sig
		logger.Sugar().Info("graceful shutdown plugin service")
		return nil
	},
}
