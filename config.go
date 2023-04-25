package p2p_database

import "github.com/caarlos0/env/v7"

type EnvConfig struct {
	PeerListenPort          int    `env:"PEER_LISTEN_PORT"`
	EthereumNetworkHost     string `env:"ETHEREUM_NETWORK_HOST,required"`
	EthereumNetworkKey      string `env:"ETHEREUM_NETWORK_KEY,required"`
	EthereumContractAddress string `env:"ETHEREUM_CONTRACT_ADDRESS,required"`
}

var config = EnvConfig{}

func init() {
	err := env.Parse(&config)
	if err != nil {
		log.Fatalf("try parse required env: %v", err)
	}
}
