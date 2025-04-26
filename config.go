package p2p_database

import (
	"github.com/caarlos0/env/v7"
	"log"
)

type Config struct {
	DisableGater     bool   `env:"DISABLE_GATER"`
	PeerListenPort   int    `env:"PEER_LISTEN_PORT"`
	WalletPrivateKey string `env:"WALLET_PRIVATE_KEY"`
	DatabaseName     string `env:"DATABASE_NAME"`
}

var EnvConfig = Config{}

func init() {
	err := env.Parse(&EnvConfig)
	if err != nil {
		log.Fatalf("try parse required env: %v", err)
	}
}
