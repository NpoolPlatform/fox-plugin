package register

import (
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/handler"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron"
)

func init() {
	mgr := handler.GetTokenMGR()
	if err := mgr.RegisterTokenInfo(tron.TronToken); err != nil {
		panic(err)
	}
}
