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
	return &EthConnectionGater{
		contract: contract,
	}
}

func (e EthConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	log.Infof("Call InterceptPeerDial %s", p.String())
	return e.checkPeerId(p, "InterceptPeerDial")
}

func (e EthConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	log.Infof("Call InterceptAddrDial %s", id.String())
	return e.checkPeerId(id, "InterceptAddrDial")
}

func (e EthConnectionGater) InterceptAccept(multiaddrs network.ConnMultiaddrs) (allow bool) {
	return true
}

func (e EthConnectionGater) InterceptSecured(direction network.Direction, id peer.ID, multiaddrs network.ConnMultiaddrs) (allow bool) {
	return true
}

func (e EthConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

func (e EthConnectionGater) checkPeerId(p peer.ID, method string) bool {
	r, _ := e.contract.ValidatePeer(p)
	return r
}
