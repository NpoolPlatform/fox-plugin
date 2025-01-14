package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NpoolPlatform/fox-plugin/pkg/config"
	"github.com/NpoolPlatform/fox-plugin/pkg/declient"
	"github.com/NpoolPlatform/fox-plugin/pkg/task"
	"github.com/NpoolPlatform/go-service-framework/pkg/logger"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"

	// register coins handler
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	_ "github.com/NpoolPlatform/fox-plugin/pkg/coins/register"
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
			Proxy:            proxyAddress,
			Position:         position,
			ConfigPath:       configPath,
			BuildChainServer: buildChainServer,
		})

		err := os.MkdirAll(logDir, 0755) //nolint
		if err != nil {
			panic(fmt.Sprintf("Fail to create log dir %v: %v", logDir, err))
		}

		err = logger.Init(logLevel, fmt.Sprintf("%v/%v.log", logDir, serviceName), zap.AddCallerSkip(1))
		if err != nil {
			panic(fmt.Sprintf("Fail to init logger: %v", err))
		}

		err = handler.GetTokenMGR().RegisterDepTokenInfosFromYaml(configPath)
		if err != nil {
			panic(fmt.Sprintf("Fail to register tokens: %v", err))
		}

		return nil
	},
	Flags: []cli.Flag{
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
		// config path
		&cli.StringFlag{
			Name:        "configPath",
			Aliases:     []string{"c"},
			Usage:       "configPath",
			EnvVars:     []string{"ENV_CONFIG_PATH"},
			Required:    false,
			Value:       "./tokens.yaml",
			Destination: &configPath,
		},
		// bc-server
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
			"run plugin Position %v",
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
