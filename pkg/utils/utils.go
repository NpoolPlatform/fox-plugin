package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NpoolPlatform/message/npool/foxproxy"
)

// ErrCoinTypeUnKnow ..
var ErrCoinTypeUnKnow = errors.New("coin type unknow")

const coinTypePrefix = "CoinType"

// ToCoinType ..
func ToCoinType(coinType string) (foxproxy.CoinType, error) {
	_coinType, ok := foxproxy.CoinType_value[fmt.Sprintf("%s%s", coinTypePrefix, coinType)]
	if !ok {
		return foxproxy.CoinType_CoinTypeUnKnow, ErrCoinTypeUnKnow
	}
	return foxproxy.CoinType(_coinType), nil
}

// nolint because CoinType not define in this package
func ToCoinName(coinType foxproxy.CoinType) string {
	coinName := strings.TrimPrefix(coinType.String(), coinTypePrefix)
	return coinName
}

func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}
