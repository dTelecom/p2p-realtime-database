package p2p_database

import (
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type EthConnectionGater struct {
	connmgr.ConnectionGater
}

func (e EthConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return checkPeerId(p)
}

func (e EthConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	return checkPeerId(id)
}

func (e EthConnectionGater) InterceptAccept(multiaddrs network.ConnMultiaddrs) (allow bool) {
	//TODO implement me
	panic("implement me")
}

func (e EthConnectionGater) InterceptSecured(direction network.Direction, id peer.ID, multiaddrs network.ConnMultiaddrs) (allow bool) {
	return checkPeerId(id)
}

func (e EthConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	//TODO implement me
	panic("implement me")
}

func checkPeerId(p peer.ID) bool {
	pubKey, err := p.ExtractPublicKey()
	if err != nil {
		return false
	}

	_, err = GetEthAddrFromPeer(pubKey)
	if err != nil {
		return false
	}

	return true
}
