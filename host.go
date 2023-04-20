package p2p_database

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	mrand "math/rand"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

func MakeHost(port int, debug bool) (host.Host, error) {
	randomGenerator := getRandomGenerator(port, debug)

	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomGenerator)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	h, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		return nil, errors.Wrap(err, "make host")
	}

	return h, nil
}

func getRandomGenerator(port int, debug bool) io.Reader {
	if debug {
		return mrand.New(mrand.NewSource(int64(port)))
	} else {
		return rand.Reader
	}
}
