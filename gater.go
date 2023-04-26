package p2p_database

import (
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type EthConnectionGater struct {
	connmgr.ConnectionGater

	contract *EthSmartContract
	logger   logging.ZapEventLogger
}

func NewEthConnectionGater(contract *EthSmartContract, logger logging.ZapEventLogger) *EthConnectionGater {
	return &EthConnectionGater{
		contract: contract,
		logger:   logger,
	}
}

func (e EthConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return e.checkPeerId(p, "InterceptPeerDial")
}

func (e EthConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	return e.checkPeerId(id, "InterceptAddrDial")
}

func (e EthConnectionGater) InterceptAccept(multiaddrs network.ConnMultiaddrs) (allow bool) {
	//todo AddrInfoFromP2pAddr from /ip4/5.16.134.127/tcp/3500 error invalid p2p multiaddr

	//a, err := peer.AddrInfoFromP2pAddr(multiaddrs.LocalMultiaddr())
	//if err != nil {
	//	e.logger.Warnf("AddrInfoFromP2pAddr from %s error %s", multiaddrs.RemoteMultiaddr(), err)
	//	return false
	//}
	//return e.checkPeerId(a.ID, "InterceptAccept")

	return true
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
		e.logger.Warnf("try validate peer %s with method %s error %s", p, method, err)
		return false
	}

	if !r {
		e.logger.Debugf("try validate peer %s with method %s: invalid", p, method)
	} else {
		e.logger.Debugf("%s peer %s validation success", method, p)
	}

	return r
}
