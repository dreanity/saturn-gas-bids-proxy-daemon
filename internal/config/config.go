package config

import (
	"encoding/hex"
	"os"

	secp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	EvmSources             []EvmSources `mapstructure:"evm_sources"`
	TreasurerPrivateKey    secp256k1.PrivKey
	TreasurerPublicKey     cryptotypes.PubKey
	SaturnNodeGrpcUrl      string `mapstructure:"saturn_node_grpc_url"`
	ChainID                string `mapstructure:"chain_id"`
	TreasurerAddress       types.AccAddress
	TreasurerRawPrivateKey string `mapstructure:"TREASURER_PRIVATE_KEY"`
}

type EvmSources struct {
	NodeUrl         string `mapstructure:"node_url"`
	ChainName       string `mapstructure:"chain_name"`
	ContractAddress string `mapstructure:"contract_address"`
}

func Init(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := parseEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func parseEnv(cfg *Config) error {
	err := godotenv.Load()
	if err != nil {
		log.Warnln("Error loading .env file")
	}

	pk := os.Getenv("TREASURER_PRIVATE_KEY")

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

	cfg.TreasurerPrivateKey = privateKey
	cfg.TreasurerPublicKey = pubKey
	cfg.TreasurerAddress = accAddress

	return nil
}
