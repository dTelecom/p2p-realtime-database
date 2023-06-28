package p2p_database

import (
	"fmt"
	"strings"

	"github.com/dTelecom/p2p-realtime-database/contracts"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/praserx/ipconv"
)

type EthSmartContract struct {
	client *contracts.Dtelecom
	logger *logging.ZapEventLogger
}

func NewEthSmartContract(conf Config, logger *logging.ZapEventLogger) (*EthSmartContract, error) {
	networkHost := conf.EthereumNetworkHost
	if !strings.HasSuffix(networkHost, "/") {
		networkHost = networkHost + "/"
	}

	client, err := ethclient.Dial(networkHost + conf.EthereumNetworkKey)
	if err != nil {
		return nil, errors.Wrap(err, "try dial eth network")
	}

	instance, err := contracts.NewDtelecom(common.HexToAddress(conf.EthereumContractAddress), client)
	if err != nil {
		return nil, errors.Wrap(err, "create smart contract instance")
	}

	return &EthSmartContract{
		client: instance,
		logger: logger,
	}, nil
}

func (e *EthSmartContract) GetEthereumClient() *contracts.Dtelecom {
	e.logger.Error("call method GetEthereumClient")
	return e.client
}

func (e *EthSmartContract) PublicKeyByAddress(address string) (string, error) {
	e.logger.Error("call method PublicKeyByAddress")
	client, err := e.client.ClientByAddress(nil, common.HexToAddress(address))
	if err != nil {
		return "", errors.Wrap(err, "try get client by address")
	}
	return client.Key, nil
}

func (e *EthSmartContract) GetBoostrapNodes() (res []peer.AddrInfo, err error) {
	e.logger.Error("call method GetBoostrapNodes")
	all, err := e.client.GetAllNode(nil)
	if err != nil {
		return nil, errors.Wrap(err, "get all nodes info")
	}

	for _, n := range all {
		ip := ipconv.IntToIPv4(uint32(n.Ip.Int64()))

		peerId, err := e.getPeerIdFromPublicKey(n.Key)
		if err != nil {
			e.logger.Errorf(
				"get bootstrap peer id %s ip %s",
				err,
				ip,
			)
			continue
		}

		e.logger.Infof("Boostrap peer from smart contract /ip4/%s/tcp/3500/p2p/%s\n", ip, peerId)

		addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/3500/p2p/%s", ip, peerId))
		if err != nil {
			e.logger.Errorf(
				"error create multiaddr bootstrap node from contract %s ip %s",
				err,
				ip,
			)
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			e.logger.Errorf("error fetch addr info for %s ip %s", err, ip)
			continue
		}

		res = append(res, *peerInfo)
	}

	if len(res) == 0 {
		e.logger.Errorf("empty list bootstrap nodes from smart contract")
	}

	return res, nil
}

func (e *EthSmartContract) ValidatePeer(p peer.ID) (bool, error) {
	ethAddr, err := e.getEthAddrFromPeer(p)
	if err != nil {
		return false, errors.Wrap(err, "get eth addr from peer")
	}

	e.logger.Error("call method NodeByAddress")
	n, err := e.client.NodeByAddress(nil, common.HexToAddress(ethAddr))
	if err != nil {
		return false, errors.Wrap(err, "fetch node by key")
	}

	return n.Active, nil
}

func (e *EthSmartContract) getEthAddrFromPeer(p peer.ID) (string, error) {
	pubkey, err := p.ExtractPublicKey()
	if err != nil {
		return "", errors.Wrap(err, "extract pub key")
	}

	dbytes, _ := pubkey.Raw()
	k, err := secp256k1.ParsePubKey(dbytes)

	if err != nil {
		return "", err
	}

	return eth_crypto.PubkeyToAddress(*k.ToECDSA()).Hex(), nil
}

func (e *EthSmartContract) getPeerIdFromPublicKey(pk string) (string, error) {
	pubKeyBytes, err := hexutil.Decode(pk)
	if err != nil {
		return "", errors.Wrap(err, "decode hex from pub key")
	}

	pubKey, err := crypto.UnmarshalSecp256k1PublicKey(pubKeyBytes)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal secp256k1 public key")
	}

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return "", errors.Wrap(err, "fetch peer_id from pubkey")
	}

	return id.String(), nil
}
