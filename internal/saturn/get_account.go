package saturn

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dreanity/saturn/app"
	"github.com/ignite/cli/ignite/pkg/cosmoscmd"
	"google.golang.org/grpc"
)

func GetAccount(grpcConn *grpc.ClientConn, address string) (*types.GenesisAccount, error) {
	queryClient := types.NewQueryClient(grpcConn)
	res, err := queryClient.Account(context.Background(), &types.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	var genAcc types.GenesisAccount

	encoding := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	cc := client.Context{}.
		WithCodec(encoding.Marshaler).
		WithInterfaceRegistry(encoding.InterfaceRegistry).
		WithTxConfig(encoding.TxConfig).
		WithLegacyAmino(encoding.Amino)

	if err := cc.Codec.UnpackAny(res.Account, &genAcc); err != nil {
		return nil, err
	}

	return &genAcc, nil
}
