package cobra

import (
	"os"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/config"
	"github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/daemon"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	ConfigPath = "config-path"
)

func InitCmd() {
	setPrefixes("saturn")
	rootCmd := &cobra.Command{
		Use:   "start",
		Short: "Start saturn daemon and set configs",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			configPath, err := cmd.Flags().GetString(ConfigPath)
			if err != nil {
				return err
			}

			cfg, err := config.Init(configPath)
			if err != nil {
				return err
			}

			if err = daemon.StartDaemon(cfg); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.Flags().StringP(ConfigPath, "c", "", "config path (required)")
	rootCmd.MarkFlagRequired(ConfigPath)

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
