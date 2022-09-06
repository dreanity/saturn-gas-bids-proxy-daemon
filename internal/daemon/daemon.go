package daemon

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/config"
	gasbidscontract "github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/gas_bids_contract"
	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/saturn"

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
	stopped := false

	go func() {
		defer func() { stop <- true }()
		for {
			select {
			case <-ticker.C:
				var paginationKey []byte
				gasBids := make([]types.GasBid, 0)

				tickerGasBid := time.NewTicker(time.Second)
				for {
					if stopped {
						tickerGasBid.Stop()
						break
					}

					<-tickerGasBid.C
					gb, pagk := getGasBids(grpcConn, paginationKey)
					log.Infoln(*gb, paginationKey)
					if gasBids == nil {
						log.Warnln("Gas bids is nil or PaginationKey is nil")
						continue
					}
					paginationKey = pagk
					gasBids = *gb

					tickerGasBid.Stop()
					break
				}

				for chainName, evmContract := range evmContracts {
					log.Infof("process %s chain", chainName)

					counterBig, err := evmContract.GetBidsCounter(nil)

					if err != nil {
						log.Errorf("Get gas bids counter error: %s", err)
						continue
					}

					counter := counterBig.Uint64()
					log.Infof("recieved %d bids count", counter)

					if counter == 0 {
						continue
					}

					var gasBid *types.GasBid

					for _, gb := range gasBids {
						if gb.Chain != chainName {
							continue
						}

						gasBid = &gb
						break
					}

					if gasBid == nil {
						gasBid = &types.GasBid{
							Chain:  chainName,
							Number: 0,
						}
					} else {
						gasBid.Number++
					}

					for i := gasBid.Number; i < counter; i++ {
						log.Infof("try request bid #%d from %s chain", i, chainName)
						bidsEvm, err := evmContract.Bids(nil, big.NewInt(int64(i)))

						if err != nil {
							log.Errorf("Bids evm contract request by index %d error: %s", i, err)
							break
						}

						log.Infof("bid #%d received from %s chain", i, chainName)
						msg := types.MsgExecuteGasBid{
							Creator:      configs.TreasurerAddress.String(),
							BidNumber:    i,
							TokenAddress: bidsEvm.PaymentAmount.String(),
							PaidAmount:   bidsEvm.PaymentAmount.String(),
							Recipient:    bidsEvm.RecipientAddr,
							Chain:        chainName,
							Scale:        uint32(bidsEvm.PaymentTokenScale),
						}

						account := getAccount(grpcConn, configs.TreasurerAddress.String())

						if account == nil {
							log.Errorf("Base account is nil")
							break
						}

						log.Infof("account received %v", account)

						err = saturn.SendExecuteGasBidMsg(
							context.Background(),
							grpcConn,
							configs.TreasurerPrivateKey,
							configs.TreasurerPublicKey,
							(*account).GetSequence(),
							(*account).GetAccountNumber(),
							configs.ChainID,
							&msg,
						)

						log.Infoln("send execute gas bid msg")

						if err != nil {
							log.Errorf("Send execute gas bid msg error: %s", err)
							break
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

	stopped = true
	ticker.Stop()
	stop <- true

	<-stop

	return nil
}
