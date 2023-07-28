package p2p_database

import (
	"sync"
	"time"

	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

type ConnectionManager struct {
	host      host.Host
	logger    *log.ZapEventLogger
	lock      sync.RWMutex
	connected map[peer.ID]multiaddr.Multiaddr
}

func NewConnectionManager(host host.Host, logger *log.ZapEventLogger) *ConnectionManager {
	m := &ConnectionManager{
		host:      host,
		logger:    logger,
		lock:      sync.RWMutex{},
		connected: map[peer.ID]multiaddr.Multiaddr{},
	}

	m.startCheckingNetwork()

	return m
}

func (c *ConnectionManager) GetPeerIdMultiAddress(id peer.ID) (multiaddr.Multiaddr, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	multiAddr, exists := c.connected[id]
	if !exists {
		return nil, errors.New("multiaddr with peer id not found")
	}

	return multiAddr, nil
}

func (c *ConnectionManager) startCheckingNetwork() {
	go func() {
		for {
			c.lock.RLock()
			connected := c.connected
			c.lock.RUnlock()

			for alreadyConnectedPeerId := range connected {
				var found bool
				for _, peerId := range c.host.Network().Peers() {
					if peerId.String() == alreadyConnectedPeerId.String() {
						found = true
					}
				}
				if !found {
					c.lock.Lock()
					c.logger.Infof("node %s disconnected\n", alreadyConnectedPeerId)
					delete(c.connected, alreadyConnectedPeerId)
					c.lock.Unlock()
				}
			}

			for _, peerId := range c.host.Network().Peers() {
				multiAddress := c.host.Peerstore().Addrs(peerId)
				if len(multiAddress) == 0 {
					c.logger.Infof("node %s has 0 multiaddress\n", peerId)
					continue
				}

				c.lock.RLock()
				_, alreadyCached := c.connected[peerId]
				c.lock.RUnlock()

				if !alreadyCached {
					c.lock.Lock()
					c.logger.Infof("node %s connected multiaddrs %s\n", peerId, multiAddress)
					c.connected[peerId] = multiAddress[0]
					c.lock.Unlock()
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()
}
