package cobra

import (
	"encoding/hex"
	"os"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	EvmWssURL              = "evm-wss-url"
	GasBidsContractAddress = "gas-bids-contaract-address"
	TreasurerPrivateKey    = "treasurer-private-key"
	SaturnNodeGrpcUrl      = "saturn-node-grpc-url"
	ChainID                = "chain-id"
)

func InitCmd() {
	setPrefixes("saturn")
	rootCmd := &cobra.Command{
		Use:   "start",
		Short: "Start saturn daemon and set configs",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			evmWssURL, err := cmd.Flags().GetString(EvmWssURL)
			if err != nil {
				return err
			}

			gasBidsContractAddress, err := cmd.Flags().GetString(GasBidsContractAddress)
			if err != nil {
				return err
			}

			pk, err := cmd.Flags().GetString(TreasurerPrivateKey)
			if err != nil {
				return err
			}

			saturnNodeGrpcUrl, err := cmd.Flags().GetString(SaturnNodeGrpcUrl)
			if err != nil {
				return err
			}

			chainId, err := cmd.Flags().GetString(ChainID)
			if err != nil {
				return err
			}

			pkBytes, err := hex.DecodeString(pk)
			if err != nil {
				return err
			}

			privateKey := secp256k1.PrivKey{Key: pkBytes}
			pubKey := privateKey.PubKey()
			accAddress, err := types.AccAddressFromHex(pubKey.Address().String())
			if err != nil {
				return err
			}

			// cfg := daemon.Configs{
			// 	PrivateKey:  privateKey,
			// 	PublicKey:   pubKey,
			// 	NodeGrpcUrl: ngu,
			// 	DrandUrls:   du,
			// 	ChainID:     cid,
			// 	Address:     accAddress,
			// }

			// if err = daemon.StartDaemon(&cfg); err != nil {
			// 	return err
			// }

			return nil
		},
	}

	rootCmd.Flags().StringP(EvmWssURL, "e", "", "Evm websocket url (required)")
	rootCmd.MarkFlagRequired(GasBidsContractAddress)
	rootCmd.Flags().StringP(GasBidsContractAddress, "g", "", "Address of the smart contract accepting gas bids (required)")
	rootCmd.MarkFlagRequired(GasBidsContractAddress)
	rootCmd.Flags().StringP(TreasurerPrivateKey, "p", "", "The treasurer private key from which the transaction will be sent (required)")
	rootCmd.MarkFlagRequired(TreasurerPrivateKey)
	rootCmd.Flags().StringP(SaturnNodeGrpcUrl, "n", "", "A saturn grpc url to the node to which the transaction will be sent (required)")
	rootCmd.MarkFlagRequired(SaturnNodeGrpcUrl)
	rootCmd.Flags().StringP(ChainID, "c", "", "Chain identifier (required)")
	rootCmd.MarkFlagRequired(ChainID)

	execute(rootCmd)
}

func execute(rootCmd *cobra.Command) {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func setPrefixes(accountAddressPrefix string) {
	// Set prefixes
	accountPubKeyPrefix := accountAddressPrefix + "pub"
	validatorAddressPrefix := accountAddressPrefix + "valoper"
	validatorPubKeyPrefix := accountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := accountAddressPrefix + "valcons"
	consNodePubKeyPrefix := accountAddressPrefix + "valconspub"

	// Set and seal config
	config := types.GetConfig()
	config.SetBech32PrefixForAccount(accountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()
}
