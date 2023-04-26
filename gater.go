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

	contract *EthSmartContract
}

func NewEthConnectionGater(contract *EthSmartContract) *EthConnectionGater {
	logging.SetLogLevel("eth-gater", "info")

	return &EthConnectionGater{
		contract: contract,
	}
}

func (e EthConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return e.checkPeerId(p, "InterceptPeerDial")
}

func (e EthConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	return e.checkPeerId(id, "InterceptAddrDial")
}

func (e EthConnectionGater) InterceptAccept(multiaddrs network.ConnMultiaddrs) (allow bool) {
	a, err := peer.AddrInfoFromP2pAddr(multiaddrs.RemoteMultiaddr())
	if err != nil {
		log.Warnf("AddrInfoFromP2pAddr from %s error %s", multiaddrs.RemoteMultiaddr(), err)
		return false
	}
	return e.checkPeerId(a.ID, "InterceptAccept")
}

func (e EthConnectionGater) InterceptSecured(direction network.Direction, id peer.ID, multiaddrs network.ConnMultiaddrs) (allow bool) {
	return e.checkPeerId(id, "InterceptSecured")
}

func (e EthConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

func (e EthConnectionGater) checkPeerId(p peer.ID, method string) bool {
	r, err := e.contract.ValidatePeer(p)

	if err != nil {
		log.Errorf("try validate peer %s with method %s error %s", p, method, err)
		return false
	}

	if !r {
		log.Warnf("try validate peer %s with method %s: invalid", p, method)
	} else {
		log.Infof("%s peer %s validation success", method, p)
	}

	return r
}
