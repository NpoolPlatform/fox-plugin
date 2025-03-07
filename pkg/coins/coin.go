package coins

import (
	"reflect"

	v1 "github.com/NpoolPlatform/message/npool/basetypes/v1"
	"github.com/NpoolPlatform/message/npool/foxproxy"
)

type TokenInfo struct {
	OfficialName        string             `yaml:"OfficialName"`
	OfficialContract    string             `yaml:"OfficialContract"`
	Contract            string             `modifiable:"true" yaml:"Contract"` // ENV is main Contract = OfficialContract
	ENV                 string             `modifiable:"true" yaml:"ENV"`      // modifiable,except mainnet
	Unit                string             `modifiable:"true" yaml:"Unit"`     // modifiable,except mainnet
	Decimal             int                `modifiable:"true" yaml:"Decimal"`  // modifiable,except mainnet
	Name                string             `modifiable:"true" yaml:"Name"`     // modifiable，if ENV is main, Name cannot be changed
	DisableRegiste      bool               `yaml:"DisableRegiste"`
	CoinType            foxproxy.CoinType  `yaml:"CoinType"`
	ChainType           foxproxy.ChainType `yaml:"ChainType"`
	ChainNativeUnit     string             `modifiable:"true" yaml:"ChainNativeUnit"`     // modifiable,except mainnet
	ChainAtomicUnit     string             `modifiable:"true" yaml:"ChainAtomicUnit"`     // modifiable,except mainnet
	ChainUnitExp        uint32             `modifiable:"true" yaml:"ChainUnitExp"`        // modifiable,except mainnet
	ChainID             string             `modifiable:"true" yaml:"ChainID"`             // modifiable,except mainnet
	ChainNickname       string             `modifiable:"true" yaml:"ChainNickname"`       // modifiable,except mainnet
	ChainNativeCoinName string             `modifiable:"true" yaml:"ChainNativeCoinName"` // modifiable,except mainnet
	GasType             v1.GasType         `yaml:"GasType"`
	BlockTime           uint16             `modifiable:"true" yaml:"BlockTime"`    // seconds;modifiable,except mainnet
	S3KeyPrxfix         string             `modifiable:"false" yaml:"S3KeyPrxfix"` // modifiable,except mainnet

	// must given from user
	LocalAPIs  []string `modifiable:"true" yaml:"LocalAPIs"`
	PublicAPIs []string `modifiable:"true" yaml:"PublicAPIs"`
}

type DepTokenInfo struct {
	TempName  string `yaml:"TempName"`
	TokenInfo `yaml:"TokenInfo"`
}

const (
	CoinNetMain = "main"
)

func GetModifiableFileds() []string {
	r := reflect.TypeOf(TokenInfo{})
	ret := []string{}
	for i := 0; i < r.NumField(); i++ {
		if r.Field(i).Tag.Get("modifiable") == "true" {
			ret = append(ret, r.Field(i).Name)
		}
	}
	return ret
}

func ToSubmitTx(tx *foxproxy.Transaction) *foxproxy.SubmitTransaction {
	return &foxproxy.SubmitTransaction{
		TransactionID: tx.TransactionID,
		Payload:       tx.Payload,
		State:         tx.State,
		LockTime:      tx.LockTime,
		ExitCode:      0,
	}
}
