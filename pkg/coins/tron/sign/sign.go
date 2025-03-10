package sign

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	addr "github.com/Geapefurit/gotron-sdk/pkg/address"
	"github.com/Geapefurit/gotron-sdk/pkg/proto/api"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"

	"google.golang.org/protobuf/proto"
)

func CreateTrxAccount(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req *foxproxy.CreateWalletRequest) (*foxproxy.CreateWalletResponse, error) {
	return CreateTronAccount(ctx, info.S3KeyPrxfix, req)
}

func CreateTronAccount(ctx context.Context, s3Strore string, req *foxproxy.CreateWalletRequest) (*foxproxy.CreateWalletResponse, error) {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	if len(priv.D.Bytes()) != 32 {
		for {
			priv, err := btcec.NewPrivateKey(btcec.S256())
			if err != nil {
				continue
			}
			if len(priv.D.Bytes()) == 32 {
				break
			}
		}
	}

	a := addr.PubkeyToAddress(priv.ToECDSA().PublicKey)
	pubkey := a.String()
	prikey := hex.EncodeToString(priv.D.Bytes())

	err = oss.PutObject(ctx, s3Strore+pubkey, []byte(prikey), true)
	if err != nil {
		return nil, err
	}

	return &foxproxy.CreateWalletResponse{Info: &foxproxy.WalletInfo{
		Address: pubkey,
	}}, nil
}

//nolint:revive
func SignTronTX(ctx context.Context, info *coins.TokenInfo, tx *foxproxy.Transaction) (*foxproxy.SubmitTransaction, error) {
	txExtension := &api.TransactionExtention{}
	err := json.Unmarshal(tx.Payload, txExtension)
	if err != nil {
		return nil, err
	}

	pk, err := oss.GetObject(ctx, info.S3KeyPrxfix+tx.From, true)
	if err != nil {
		return nil, err
	}

	privateBytes, err := hex.DecodeString(string(pk))
	if err != nil {
		return nil, err
	}
	priv := crypto.ToECDSAUnsafe(privateBytes)
	rawData, err := proto.Marshal(txExtension.Transaction.GetRawData())
	if err != nil {
		return nil, fmt.Errorf("proto marshal tx raw data error: %v", err)
	}

	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)

	signature, err := crypto.Sign(hash, priv)
	if err != nil {
		return nil, fmt.Errorf("sign error: %v", err)
	}

	txExtension.Transaction.Signature = append(txExtension.Transaction.Signature, signature)
	payload, err := json.Marshal(txExtension)
	if err != nil {
		return nil, err
	}

	submitTx := coins.ToSubmitTx(tx)
	submitTx.Payload = payload

	return submitTx, nil
}
