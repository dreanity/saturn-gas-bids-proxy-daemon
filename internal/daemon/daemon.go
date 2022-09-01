package daemon

import (
	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/config"
	"google.golang.org/grpc"
)

func StartDaemon(configs *config.Config) error {
	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		configs.SaturnNodeGrpcUrl, // Or your gRPC server address.
		grpc.WithInsecure(),       // The Cosmos SDK doesn't support any transport security mechanism.
	)
	if err != nil {
		return err
	}
	defer grpcConn.Close()

	return nil

}

func getSaturnGasBid() {}
