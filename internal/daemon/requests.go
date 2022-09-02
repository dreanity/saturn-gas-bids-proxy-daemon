package daemon

import (
	"time"

	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/saturn"
	"github.com/dreanity/saturn/x/treasury/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

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
