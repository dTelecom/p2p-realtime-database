package p2p_database

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"

	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

func MakeHost(ctx context.Context, port int, debug bool) (host.Host, *dual.DHT, error) {
	//randomGenerator := getRandomGenerator(port, debug)

	//prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomGenerator)
	//if err != nil {
	//	log.Println(err)
	//	return nil, nil, err
	//}

	prvKey, err := eth_crypto.HexToECDSA("9ad26aeb690b637e7bce2718a08f25fe11ea79142fd1f6987b0b2cfef8ab76ae")
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
	opts = append(opts, libp2p.ConnectionGater(EthConnectionGater{}))

	return ipfslite.SetupLibp2p(
		ctx,
		priv,
		nil,
		[]multiaddr.Multiaddr{sourceMultiAddr},
		nil,
		ipfslite.Libp2pOptionsExtra...,
	)
}

func getRandomGenerator(port int, debug bool) io.Reader {
	if debug {
		return mrand.New(mrand.NewSource(int64(port)))
	} else {
		return rand.Reader
	}
}
