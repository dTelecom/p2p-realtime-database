package p2p_database

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("eth-gater")

type EthConnectionGater struct {
	connmgr.ConnectionGater
}

func NewEthConnectionGater() *EthConnectionGater {
	return &EthConnectionGater{}
}

func (e EthConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return checkPeerId(p, "InterceptPeerDial")
}

func (e EthConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	return checkPeerId(id, "InterceptAddrDial")
}

func (e EthConnectionGater) InterceptAccept(multiaddrs network.ConnMultiaddrs) (allow bool) {
	return true
}

func (e EthConnectionGater) InterceptSecured(direction network.Direction, id peer.ID, multiaddrs network.ConnMultiaddrs) (allow bool) {
	return checkPeerId(id, "InterceptSecured")
}

func (e EthConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

func checkPeerId(p peer.ID, method string) bool {
	pubKey, err := p.ExtractPublicKey()
	if err != nil {
		log.Errorf("method %s gater cannot extract public key: %s of %s", method, err, p.String())
		return false
	}

	_, err = GetEthAddrFromPeer(pubKey)
	if err != nil {
		log.Errorf("method %s error extract eth address: %s of %s", method, err, p.String())
		return false
	}

	return true
}
