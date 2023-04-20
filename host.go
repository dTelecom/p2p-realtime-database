package p2p_database

import (
	"context"
	"crypto/rand"
	"fmt"
	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"io"
	"log"
	mrand "math/rand"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

func MakeHost(ctx context.Context, port int, debug bool) (host.Host, *dual.DHT, error) {
	randomGenerator := getRandomGenerator(port, debug)

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomGenerator)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	return ipfslite.SetupLibp2p(
		ctx,
		prvKey,
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
