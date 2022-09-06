package saturn

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/tx"
	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/dreanity/saturn/app"
	"github.com/dreanity/saturn/x/treasury/types"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	uhydrogen = "uhydrogen"
	gasLimit  = 150_000
)

func SendExecuteGasBidMsg(
	ctx context.Context,
	grpcConn *grpc.ClientConn,
	privKey secp256k1.PrivKey,
	pubKey cryptotypes.PubKey,
	accSeq uint64,
	accNum uint64,
	chainId string,
	msg *types.MsgExecuteGasBid,
) error {
	encoding := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	txBuilder := encoding.TxConfig.NewTxBuilder()

	if err := txBuilder.SetMsgs(msg); err != nil {
		return err
	}

	feeAmount := sdktypes.NewCoin(uhydrogen, sdktypes.NewInt(gasLimit))

	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(sdktypes.NewCoins(feeAmount))

	// -------------------------------------------------------------------

	sigV2 := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  encoding.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: accSeq,
	}

	txBuilder.SetSignatures(sigV2)

	signerData := xauthsigning.SignerData{
		ChainID:       chainId,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}

	sigV2, err := tx.SignWithPrivKey(
		encoding.TxConfig.SignModeHandler().DefaultMode(),
		signerData,
		txBuilder,
		&privKey,
		encoding.TxConfig,
		accSeq)

	if err != nil {
		return err
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return err
	}

	txSender := sdktx.NewServiceClient(grpcConn)
	txBytes, err := encoding.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}

	res, err := txSender.BroadcastTx(
		ctx,
		&sdktx.BroadcastTxRequest{
			Mode:    sdktx.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		})

	if err != nil {
		return err
	}

	log.Infoln("response send execute gas: %v", res)

	return nil
}
