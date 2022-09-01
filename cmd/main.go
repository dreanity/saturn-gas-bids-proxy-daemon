package main

import (
	"github.com/dreanity/saturn-gas-bids-proxy-daemon/cmd/cobra"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(new(log.TextFormatter))
	cobra.InitCmd()
}
