package p2p_database

import (
	"fmt"
	"github.com/dTelecom/p2p-database/contracts"
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
	"strings"
)

var logger = logging.Logger("eth-gater")

type EthSmartContract struct {
	client *contracts.Dtelecom
}

func NewEthSmartContract() (*EthSmartContract, error) {
	networkHost := config.EthereumNetworkHost
	if !strings.HasSuffix(networkHost, "/") {
		networkHost = networkHost + "/"
	}

	client, err := ethclient.Dial(networkHost + config.EthereumNetworkKey)
	if err != nil {
		return nil, errors.Wrap(err, "try dial eth network")
	}

	instance, err := contracts.NewDtelecom(common.HexToAddress(config.EthereumContractAddress), client)
	if err != nil {
		return nil, errors.Wrap(err, "create smart contract instance")
	}

	return &EthSmartContract{
		client: instance,
	}, nil
}

func (e *EthSmartContract) GetBoostrapNodes() (res []peer.AddrInfo, err error) {
	all, err := e.client.GetAllNode(nil)
	if err != nil {
		return nil, errors.Wrap(err, "get all nodes info")
	}

	for _, n := range all {
		ip := ipconv.IntToIPv4(uint32(n.Ip.Int64()))

		peerId, err := e.getPeerIdFromPublicKey(n.Key)
		if err != nil {
			logger.Errorf(
				"get bootstrap peer id %s ip %s",
				err,
				ip,
			)
			continue
		}

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
		log.Errorf("empty list bootstrap nodes from smart contract")
	}

	return res, nil
}

func (e *EthSmartContract) ValidatePeer(p peer.ID) (bool, error) {
	ethAddr, err := e.getEthAddrFromPeer(p)
	if err != nil {
		return false, errors.Wrap(err, "get eth addr from peer")
	}

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
