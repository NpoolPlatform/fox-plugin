package main

import (
	"fmt"

	cli "github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	// register coins handler
	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	_ "github.com/NpoolPlatform/fox-plugin/pkg/coins/register"
	"github.com/NpoolPlatform/fox-plugin/pkg/utils"
)

var listCmd = &cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "List All token info",
	Before: func(ctx *cli.Context) error {
		return nil
	},
	Action: func(c *cli.Context) error {
		fmt.Println("################### All Token Infos start ####################")
		fmt.Println(GetAllTokenInfos())
		fmt.Println("")
		fmt.Println("################### Modifyable fields for token ##############")
		fmt.Println(utils.PrettyStruct(coins.GetModifiableFileds()))
		fmt.Println("Tips:")
		fmt.Println("(1) If ENV is main, only [LocalAPIs PublicAPIs] can be modify")
		fmt.Println("(2) The unit of BlockTime is second")
		fmt.Println("(3) LocalAPIs and PublicAPIs must be provied by deployer")
		return nil
	},
}

func GetAllTokenInfos() string {
	_infos := handler.GetTokenMGR().GetTokenInfos()
	tokenInfos := struct {
		Infos []*coins.DepTokenInfo `yaml:"TokenInfos"`
	}{}
	for _, v := range _infos {
		tokenInfos.Infos = append(tokenInfos.Infos, &coins.DepTokenInfo{
			TempName:  v.Name,
			TokenInfo: *v,
		})
	}
	out, err := yaml.Marshal(tokenInfos)
	if err != nil {
		panic(err)
	}
	return string(out)
}
