package daemon

import (
	"strings"

	gasbidscontract "github.com/dreanity/saturn-gas-bids-proxy-daemon/internal/gas_bids_contract"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
)

const key_store = `{"version":3,"id":"9a0483f4-9d3a-445a-a4d6-5621ecbe2ccd","address":"8d341a611dbab474ef7c7efda9ac360867dd92d0","crypto":{"ciphertext":"05956c711d2b18c40e9dfba88814515b13d8de992df2355b2d24d97dd7931d6e","cipherparams":{"iv":"f1f3b9a444b3577a9c5e0ee666ead6c3"},"cipher":"aes-128-ctr","kdf":"scrypt","kdfparams":{"dklen":32,"salt":"0adb4e7e12e295bb535142b9005905147d55e921c7bcec6424af37fe1b180c76","n":262144,"r":8,"p":1},"mac":"d9b03091b3ddf672de591a07af6ced18da9ba2a35947ce24bb8054be288334e0"}}`
const USDT_ADDR = `0x337610d27c682E347C9cD60BD4b3b107C9d34dDd`
const RECIPIENT = `saturn1wvvyvp4y7uq3sdsaz2vwslh8kk8z08jsvp0je9`

func testCreateBid() error {
	client, err := ethclient.Dial("https://data-seed-prebsc-1-s1.binance.org:8545/")
	if err != nil {
		return err
	}

	address := common.HexToAddress("0x3C957b9dc20d541a277F9546bccbe2f8EEEDd986")
	contract, err := gasbidscontract.NewGasbidscontract(address, client)
	if err != nil {
		return err
	}

	transactor, err := bind.NewTransactor(strings.NewReader(key_store), "strong_password")
	if err != nil {
		return err
	}

	tx, err := contract.CreateBid(transactor, common.HexToAddress(USDT_ADDR), RECIPIENT)
	if err != nil {
		return err
	}

	log.Infoln(tx)

	return nil
}
