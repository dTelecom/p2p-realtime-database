package p2p_database

import (
	"fmt"
	"sync"

	"github.com/dTelecom/p2p-realtime-database/internal/common"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

type SolanaConnectionGater struct {
	cache  sync.Map
	logger *logging.ZapEventLogger
}

func NewSolanaConnectionGater(logger *logging.ZapEventLogger) *SolanaConnectionGater {
	g := &SolanaConnectionGater{
		logger: logger,
	}
	return g
}

func (e *SolanaConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	return e.checkPeerId(p, "InterceptPeerDial")
}

func (e *SolanaConnectionGater) InterceptAddrDial(id peer.ID, multiaddr multiaddr.Multiaddr) (allow bool) {
	return e.checkPeerId(id, "InterceptAddrDial")
}

func (e *SolanaConnectionGater) InterceptAccept(multiaddrs network.ConnMultiaddrs) (allow bool) {
	return true
}

func (e *SolanaConnectionGater) InterceptSecured(direction network.Direction, id peer.ID, multiaddrs network.ConnMultiaddrs) (allow bool) {
	return e.checkPeerId(id, "InterceptSecured")
}

func (e *SolanaConnectionGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}

func (e *SolanaConnectionGater) checkPeerId(p peer.ID, method string) bool {
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
	r, err := ValidatePeer(p, e.logger)

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

type Node struct {
	Ip  string
	Key string
}

func GetBoostrapNodes(logger *logging.ZapEventLogger) (res []peer.AddrInfo, err error) {
	all := []Node{
		{
			Ip:  "34.175.243.9",
			Key: "5g3euBKXqhdbfzkgbWQ7o1C6HQzbyr1noX6wiqfv2i3x",
		},
		{
			Ip:  "35.205.115.103",
			Key: "644PeDtMTPE1WFSaTC8BSjaNUN697frNRifciWSqiAZz",
		},
		{
			Ip:  "34.165.46.203",
			Key: "CiKDiqBjHs9jTaJANgreQkVgA6J4YhVbYx55tU7swKfk",
		},
	}

	for _, n := range all {
		ip := n.Ip

		peerId, err := getPeerIdFromPublicKey(n.Key)
		if err != nil {
			logger.Errorf(
				"get bootstrap peer id %s ip %s",
				err,
				ip,
			)
			continue
		}

		logger.Infof("Boostrap peer from smart contract /ip4/%s/tcp/3500/p2p/%s\n", ip, peerId)

		addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/3500/p2p/%s", ip, peerId))
		if err != nil {
			logger.Errorf(
				"error create multiaddr bootstrap node from contract %s ip %s",
				err,
				ip,
			)
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			logger.Errorf("error fetch addr info for %s ip %s", err, ip)
			continue
		}

		res = append(res, *peerInfo)
	}

	if len(res) == 0 {
		logger.Errorf("empty list bootstrap nodes from smart contract")
	}

	return res, nil
}

func ValidatePeer(p peer.ID, logger *logging.ZapEventLogger) (bool, error) {
	solanaAddr, err := getSolanaAddrFromPeer(p)
	if err != nil {
		return false, errors.Wrap(err, "get solana addr from peer")
	}

	// Here you would normally validate against a smart contract or other authority
	// For now, we allow all valid Solana-derived peer IDs
	logger.Debugf("Validating Solana peer with address %s", solanaAddr)

	all := []Node{
		{
			Ip:  "34.175.243.9",
			Key: "5g3euBKXqhdbfzkgbWQ7o1C6HQzbyr1noX6wiqfv2i3x",
		},
		{
			Ip:  "35.205.115.103",
			Key: "644PeDtMTPE1WFSaTC8BSjaNUN697frNRifciWSqiAZz",
		},
		{
			Ip:  "34.165.46.203",
			Key: "CiKDiqBjHs9jTaJANgreQkVgA6J4YhVbYx55tU7swKfk",
		},
	}

	for _, n := range all {
		if n.Key == solanaAddr {
			return true, nil
		}
	}

	// Allow all peers for testing
	return true, nil
}

func getSolanaAddrFromPeer(p peer.ID) (string, error) {
	pubkey, err := p.ExtractPublicKey()
	if err != nil {
		return "", errors.Wrap(err, "extract pub key")
	}

	// For Ed25519 keys, the public key is directly usable as a Solana address
	dbytes, err := pubkey.Raw()
	if err != nil {
		return "", errors.Wrap(err, "extract raw bytes from public key")
	}

	// Return the Solana base58 encoded public key (address)
	return base58.Encode(dbytes), nil
}

func getPeerIdFromPublicKey(pk string) (string, error) {
	// Decode the base58-encoded Solana public key
	pubKeyBytes, err := base58.Decode(pk)
	if err != nil {
		return "", errors.Wrap(err, "decode base58 solana public key")
	}

	// Unmarshal as Ed25519 public key
	pubKey, err := crypto.UnmarshalEd25519PublicKey(pubKeyBytes)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal Ed25519 public key")
	}

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return "", errors.Wrap(err, "fetch peer_id from pubkey")
	}

	return id.String(), nil
}

// NewSolanaConnectionGaterWithCommonLogger creates a new SolanaConnectionGater with a common.Logger
func NewSolanaConnectionGaterWithCommonLogger(logger common.Logger) *SolanaConnectionGater {
	// Use the standard IPFS logger for internal gater operations
	zapLogger := logging.Logger("solana-gater")
	return NewSolanaConnectionGater(zapLogger)
}

// GetBoostrapNodesWithCommonLogger gets bootstrap nodes with a common.Logger
func GetBoostrapNodesWithCommonLogger(logger common.Logger) ([]peer.AddrInfo, error) {
	// Use the standard IPFS logger for internal bootstrap operations
	zapLogger := logging.Logger("bootstrap")
	return GetBoostrapNodes(zapLogger)
}
