package p2p_database

import (
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

const (
	cacheTTL = 30 * time.Second
)

type EthConnectionGater struct {
	connmgr.ConnectionGater

	alreadyConnected map[peer.ID]struct{}

	lock  sync.Mutex
	cache map[string]map[peer.ID]bool

	contract *EthSmartContract
	logger   logging.ZapEventLogger
}

func NewEthConnectionGater(contract *EthSmartContract, alreadyConnected map[peer.ID]struct{}, logger logging.ZapEventLogger) *EthConnectionGater {
	g := &EthConnectionGater{
		contract:         contract,
		lock:             sync.Mutex{},
		cache:            make(map[string]map[peer.ID]bool),
		alreadyConnected: alreadyConnected,
		logger:           logger,
	}

	g.startCleanupCacheProcess()

	return g
}

func (e *EthConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return e.checkPeerId(p, "InterceptPeerDial")
}

func (e *EthConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	e.logger.Errorf("InterceptAddrDial %s %s", id, multiaddr.String())
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

	_, alreadyConnect := e.alreadyConnected[p]
	if alreadyConnect {
		e.logger.Errorf("try validate peer %s with method %s error %s", p, method, errors.New("peer with same id already connected"))
		return false
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	cache, ok := e.cache[method]
	if ok {
		result, ok := cache[p]
		if ok {
			return result
		}
	} else {
		e.cache[method] = map[peer.ID]bool{}
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

	e.cache[method][p] = r

	return r
}

func (e *EthConnectionGater) startCleanupCacheProcess() {
	go func() {
		ticker := time.NewTicker(cacheTTL)
		for {
			e.logger.Debugf("cleanup cache gater len %d", len(e.cache))

			e.lock.Lock()
			e.cache = make(map[string]map[peer.ID]bool)
			e.lock.Unlock()

			<-ticker.C
		}
	}()
}
