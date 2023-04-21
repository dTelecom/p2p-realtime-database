package p2p_database

import (
	"context"
	"fmt"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

func MakeHost(ctx context.Context, ethPrivateKey string, port int, debug bool) (host.Host, *dual.DHT, error) {
	prvKey, err := eth_crypto.HexToECDSA(ethPrivateKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "hex to ecdsa eth private key")
	}

	privKeyBytes := eth_crypto.FromECDSA(prvKey)
	priv, err := crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "UnmarshalSecp256k1PrivateKey from eth private key")
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	opts := ipfslite.Libp2pOptionsExtra
	opts = append(
		opts,
		libp2p.ConnectionGater(NewEthConnectionGater()),
	)

	return ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{sourceMultiAddr},
		nil,
		opts...,
	)
}
