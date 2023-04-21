package p2p_database

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func GetEthAddrFromPeer(pubkey crypto.PubKey) (string, error) {
	dbytes, _ := pubkey.Raw()
	k, err := secp256k1.ParsePubKey(dbytes)

	if err != nil {
		return "", err
	}

	return eth_crypto.PubkeyToAddress(*k.ToECDSA()).Hex(), nil
}
