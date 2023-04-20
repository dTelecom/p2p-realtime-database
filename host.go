package p2p_database

import (
	"crypto/rand"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"io"
	"log"
	mrand "math/rand"
)

func MakeHost(port int, debug bool) (host.Host, crypto.PrivKey, crypto.PubKey, error) {
	randomGenerator := getRandomGenerator(port, debug)

	prvKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, randomGenerator)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, err
	}

	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	h, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "make host")
	}

	return h, prvKey, pubKey, nil
}

func getRandomGenerator(port int, debug bool) io.Reader {
	if debug {
		return mrand.New(mrand.NewSource(int64(port)))
	} else {
		return rand.Reader
	}
}
