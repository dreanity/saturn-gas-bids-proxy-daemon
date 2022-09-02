package daemon

import (
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/config"
	gasbidscontract "github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/gas_bids_contract"

	"github.com/dreanity/saturn/x/treasury/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
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

	evmClients := make(map[string]*ethclient.Client)
	evmContracts := make(map[string]*gasbidscontract.Gasbidscontract)

	for _, evmSource := range configs.EvmSources {
		client, err := ethclient.Dial(evmSource.NodeUrl)
		if err != nil {
			return err
		}
		evmClients[evmSource.ChainName] = client

		address := common.HexToAddress(evmSource.ContractAddress)
		contract, err := gasbidscontract.NewGasbidscontract(address, client)
		if err != nil {
			return nil
		}

		evmContracts[evmSource.ChainName] = contract
	}

	ticker := time.NewTicker(time.Second)
	stop := make(chan bool)

	go func() {
		defer func() { stop <- true }()
		for {
			select {
			case <-ticker.C:
				var paginationKey []byte
				var gasBids *[]types.GasBid

				tickerGasBid := time.NewTicker(time.Second)

				for {
					<-tickerGasBid.C
					gasBids, paginationKey = getGasBids(grpcConn, paginationKey)
					log.Infoln(gasBids, paginationKey)
					if gasBids == nil {
						log.Warnln("Gas bids is nil or PaginationKey is nil")
						continue
					}

					ticker.Stop()
					break
				}

				for chainName, evmContract := range evmContracts {
					counterBig, err := evmContract.GetBidsCounter(nil)
					counter := counterBig.Uint64()
					if err != nil {
						log.Errorf("Get gas bids counter error: %s", err)
						continue
					}

					if counter == 0 {
						continue
					}

					var gasBid *types.GasBid

					for _, gb := range *gasBids {
						if gb.FromChain != chainName {
							continue
						}

						gasBid = &gb
						break
					}

					if gasBid == nil {
						gasBid = &types.GasBid{
							FromChain: chainName,
							Number:    0,
						}
					} else {
						gasBid.Number++
					}

					for i := gasBid.Number; i < counter; i++ {
						bidsEvm, err := evmContract.Bids(nil, big.NewInt(int64(i)))
						if err != nil {
							log.Errorf("Bids evm contract request by index %d error: %s", i, err)
							break
						}

						msg := types.MsgExecuteGasBid{
							Creator:    configs.TreasurerAddress.String(),
							BidNumber:  i,
							Currency:   bidsEvm.PaymentTokenAddr.String(),
							PaidAmount: bidsEvm.PaymentAmount.String(),
							Recipient:  bidsEvm.RecipientAddr,
							FromChain:  chainName,
						}
					}

				}
			case <-stop:
				log.Info("Stopping the deamon")
				return
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ticker.Stop()

	stop <- true

	<-stop

	return nil
}
