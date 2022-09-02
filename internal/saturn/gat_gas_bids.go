package saturn

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dreanity/saturn/x/treasury/types"
	"google.golang.org/grpc"
)

func GetGasBids(grpcConn *grpc.ClientConn, paginationKey []byte) (*types.QueryAllGasBidResponse, error) {
	queryClient := types.NewQueryClient(grpcConn)
	res, err := queryClient.GasBidAll(context.Background(), &types.QueryAllGasBidRequest{
		Pagination: &query.PageRequest{
			Key: paginationKey,
		},
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
