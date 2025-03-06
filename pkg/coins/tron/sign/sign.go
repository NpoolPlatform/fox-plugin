package sign

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	addr "github.com/Geapefurit/gotron-sdk/pkg/address"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins"
	"github.com/NpoolPlatform/fox-plugin/pkg/coins/tron"
	"github.com/NpoolPlatform/go-service-framework/pkg/oss"
	"github.com/NpoolPlatform/message/npool/foxproxy"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"

	"google.golang.org/protobuf/proto"
)

func SignTrxMSG(ctx context.Context, in []byte, info *coins.TokenInfo) (out []byte, err error) {
	return SignTronMSG(ctx, info.S3KeyPrxfix, in)
}

func CreateTrxAccount(ctx context.Context, coinInfo *foxproxy.CoinInfo, info *coins.TokenInfo, req *foxproxy.CreateWalletRequest) (*foxproxy.CreateWalletResponse, error) {
	return CreateTronAccount(ctx, info.S3KeyPrxfix, req)
}

func SignTronMSG(ctx context.Context, s3Strore string, in []byte) (out []byte, err error) {
	signMsgTx := &tron.SignMsgTx{}
	err = json.Unmarshal(in, signMsgTx)
	if err != nil {
		return in, err
	}
	if signMsgTx.TxExtension == nil {
		return in, errors.New(tron.BuildTransactionFailed)
	}

	pk, err := oss.GetObject(ctx, s3Strore+signMsgTx.Base.From, true)
	if err != nil {
		return in, err
	}

	privateBytes, err := hex.DecodeString(string(pk))
	if err != nil {
		return in, err
	}
	transaction := signMsgTx.TxExtension.Transaction
	priv := crypto.ToECDSAUnsafe(privateBytes)
	rawData, err := proto.Marshal(transaction.GetRawData())
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

	transaction.Signature = append(transaction.Signature, signature)
	signedMsg := &tron.BroadcastRequest{TxExtension: signMsgTx.TxExtension}
	return json.Marshal(signedMsg)
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
