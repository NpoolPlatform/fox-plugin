package sol

import (
	"errors"
	"math/big"
	"strings"

	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	v1 "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	solana "github.com/gagliardetto/solana-go"
)

var (
	// EmptyWalletL ..
	EmptyWalletL = big.Int{}
	// EmptyWalletS ..
	EmptyWalletS = big.Float{}
)

const (
	ChainType           = foxproxy.ChainType_Solana
	ChainNativeUnit     = "SOL"
	ChainAtomicUnit     = "lamport"
	ChainUnitExp        = 9
	ChainNativeCoinName = "solana"
	ChainID             = "101"
)

var (
	// ErrSolBlockNotFound ..
	ErrSolBlockNotFound = errors.New("not found confirmed block in solana chain")
	// ErrSolSignatureWrong ..
	ErrSolSignatureWrong = errors.New("solana signature is wrong or failed")
)

var (
	SolTransactionFailed = `sol transaction failed`
	lamportsLow          = `Transfer: insufficient lamports`
	txFailed             = `Transaction simulation failed`
	txSignatureWrong     = `Transaction signature verification failure`
	txSignatureNotMatch  = `There is a mismatch in the length of the transaction signature`
	txVersionWrong       = `Transaction version (0) is not supported by the requesting client`
	stopErrMsg           = []string{lamportsLow, SolTransactionFailed, txFailed, txSignatureWrong, txSignatureNotMatch, txVersionWrong}

	SolanaToken = &coins.TokenInfo{
		OfficialName:        "Solana",
		OfficialContract:    ChainNativeCoinName,
		Contract:            ChainNativeCoinName,
		ENV:                 coins.CoinNetMain,
		Unit:                "SOL",
		Decimal:             9,
		Name:                ChainNativeCoinName,
		DisableRegiste:      false,
		CoinType:            foxproxy.CoinType_CoinTypesolana,
		ChainType:           ChainType,
		ChainNativeUnit:     ChainNativeUnit,
		ChainAtomicUnit:     ChainAtomicUnit,
		ChainUnitExp:        ChainUnitExp,
		ChainID:             ChainID,
		ChainNickname:       ChainType.String(),
		ChainNativeCoinName: ChainNativeCoinName,
		GasType:             v1.GasType_GasUnsupported,
		BlockTime:           1,
		S3KeyPrxfix:         "solana/",
	}
)

func ToSol(larm uint64) *big.Float {
	// Convert lamports to sol:
	return big.NewFloat(0).
		Quo(
			big.NewFloat(0).SetUint64(larm),
			big.NewFloat(0).SetUint64(solana.LAMPORTS_PER_SOL),
		)
}

func ToLarm(value float64) (uint64, big.Accuracy) {
	return big.NewFloat(0).Mul(
		big.NewFloat(0).SetFloat64(value),
		big.NewFloat(0).SetUint64(solana.LAMPORTS_PER_SOL),
	).Uint64()
}

func TxFailErr(err error) bool {
	if err == nil {
		return false
	}

	for _, v := range stopErrMsg {
		if strings.Contains(err.Error(), v) {
			return true
		}
	}
	return false
}
