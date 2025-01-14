package main

import (
	"fmt"
	"log"
	"os"

	"github.com/NpoolPlatform/go-service-framework/pkg/version"
	banner "github.com/common-nighthawk/go-figure"
	cli "github.com/urfave/cli/v2"
)

const (
	serviceName = "fox Plugin"
	usageText   = "fox Plugin Service"
)

var (
	proxyAddress     string
	logDir           string
	logLevel         string
	position         string
	configPath       string
	buildChainServer string
)

func main() {
	commands := cli.Commands{runCmd, listCmd}
	description := fmt.Sprintf(
		"%v service cli\nFor help on any individual command run <%v COMMAND -h>\n",
		serviceName,
		serviceName,
	)

	banner.NewColorFigure(serviceName, "", "green", true).Print()
	vsion, err := version.GetVersion()
	if err != nil {
		panic(fmt.Sprintf("fail to get version: %v", err))
	}

	app := &cli.App{
		Name:        serviceName,
		Version:     vsion,
		Description: description,
		Usage:       usageText,
		Commands:    commands,
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatalf("fail to run %v: %v", serviceName, err)
	}
}
