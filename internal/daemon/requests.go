package daemon

import (
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/saturn"
	"github.com/dreanity/saturn/x/treasury/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func getAccount(grpcConn *grpc.ClientConn, address string) *authtypes.GenesisAccount {
	accountChan := make(chan *authtypes.GenesisAccount)

	go func() {
		account, err := saturn.GetAccount(grpcConn, address)

		if err != nil {
			log.Error(err)
			accountChan <- nil
			return
		}

		accountChan <- account
	}()

	select {
	case account := <-accountChan:
		return account
	case <-time.After(2 * time.Second):
		log.Warn("The base account request time has expired")
		return nil
	}
}

func getGasBids(grpcConn *grpc.ClientConn, paginationKey []byte) (*[]types.GasBid, []byte) {
	gasBidsChan := make(chan *types.QueryAllGasBidResponse)

	go func() {
		res, err := saturn.GetGasBids(grpcConn, paginationKey)
		if err != nil {
			log.Errorf("Get gas bids error: %s", err)
			gasBidsChan <- nil
		}
		gasBidsChan <- res
	}()

	select {
	case gasBids := <-gasBidsChan:
		if gasBids != nil {
			return &gasBids.GasBid, gasBids.Pagination.NextKey
		}

		return nil, nil
	case <-time.After(2 * time.Second):
		log.Warn("The gas bids request time has expired")
		return nil, nil
	}
}
