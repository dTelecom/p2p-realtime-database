package p2p_database

import (
	"sync"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type EthConnectionGater struct {
	cache sync.Map
	contract *EthSmartContract
	logger   *logging.ZapEventLogger
}

func NewEthConnectionGater(contract *EthSmartContract, logger *logging.ZapEventLogger) *EthConnectionGater {
	g := &EthConnectionGater{
		contract:          contract,
		logger:            logger,
	}

	return g
}

func (e *EthConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return e.checkPeerId(p, "InterceptPeerDial")
}

func (e *EthConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	return e.checkPeerId(id, "InterceptAddrDial")
}

func (e *EthConnectionGater) InterceptAccept(multiaddrs network.ConnMultiaddrs) (allow bool) {
	return true
}

func (e *EthConnectionGater) InterceptSecured(direction network.Direction, id peer.ID, multiaddrs network.ConnMultiaddrs) (allow bool) {
	return e.checkPeerId(id, "InterceptSecured")
}

func (e *EthConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

func (e *EthConnectionGater) checkPeerId(p peer.ID, method string) bool {
	if EnvConfig.DisableGater {
		return true
	}

	cachedRaw, ok := e.cache.Load(p)
	if ok {
		cached, ok := cachedRaw.(bool)
		if ok {
			return cached
		}
	}

	e.logger.Debugf("call method %s with %s", method, p)
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

	e.cache.Store(p, r)

	return r
}
