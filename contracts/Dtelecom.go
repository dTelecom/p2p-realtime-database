// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// DtelecomNode is an auto generated low-level Go binding around an user-defined struct.
type DtelecomNode struct {
	Ip     *big.Int
	Active bool
	Key    string
}

// DtelecomMetaData contains all meta data concerning the Dtelecom contract.
var DtelecomMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"clientByAddress\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"until\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"nodeByAddress\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"ip\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"ip\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"name\":\"addNode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"removeNodeByAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllNode\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"ip\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"internalType\":\"structDtelecom.Node[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"limit\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"until\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"key\",\"type\":\"string\"}],\"name\":\"addClient\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"removeClientByAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DtelecomABI is the input ABI used to generate the binding from.
// Deprecated: Use DtelecomMetaData.ABI instead.
var DtelecomABI = DtelecomMetaData.ABI

// Dtelecom is an auto generated Go binding around an Ethereum contract.
type Dtelecom struct {
	DtelecomCaller     // Read-only binding to the contract
	DtelecomTransactor // Write-only binding to the contract
	DtelecomFilterer   // Log filterer for contract events
}

// DtelecomCaller is an auto generated read-only Go binding around an Ethereum contract.
type DtelecomCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DtelecomTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DtelecomTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DtelecomFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DtelecomFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DtelecomSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DtelecomSession struct {
	Contract     *Dtelecom         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DtelecomCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DtelecomCallerSession struct {
	Contract *DtelecomCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// DtelecomTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DtelecomTransactorSession struct {
	Contract     *DtelecomTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// DtelecomRaw is an auto generated low-level Go binding around an Ethereum contract.
type DtelecomRaw struct {
	Contract *Dtelecom // Generic contract binding to access the raw methods on
}

// DtelecomCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DtelecomCallerRaw struct {
	Contract *DtelecomCaller // Generic read-only contract binding to access the raw methods on
}

// DtelecomTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DtelecomTransactorRaw struct {
	Contract *DtelecomTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDtelecom creates a new instance of Dtelecom, bound to a specific deployed contract.
func NewDtelecom(address common.Address, backend bind.ContractBackend) (*Dtelecom, error) {
	contract, err := bindDtelecom(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Dtelecom{DtelecomCaller: DtelecomCaller{contract: contract}, DtelecomTransactor: DtelecomTransactor{contract: contract}, DtelecomFilterer: DtelecomFilterer{contract: contract}}, nil
}

// NewDtelecomCaller creates a new read-only instance of Dtelecom, bound to a specific deployed contract.
func NewDtelecomCaller(address common.Address, caller bind.ContractCaller) (*DtelecomCaller, error) {
	contract, err := bindDtelecom(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DtelecomCaller{contract: contract}, nil
}

// NewDtelecomTransactor creates a new write-only instance of Dtelecom, bound to a specific deployed contract.
func NewDtelecomTransactor(address common.Address, transactor bind.ContractTransactor) (*DtelecomTransactor, error) {
	contract, err := bindDtelecom(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DtelecomTransactor{contract: contract}, nil
}

// NewDtelecomFilterer creates a new log filterer instance of Dtelecom, bound to a specific deployed contract.
func NewDtelecomFilterer(address common.Address, filterer bind.ContractFilterer) (*DtelecomFilterer, error) {
	contract, err := bindDtelecom(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DtelecomFilterer{contract: contract}, nil
}

// bindDtelecom binds a generic wrapper to an already deployed contract.
func bindDtelecom(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DtelecomABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Dtelecom *DtelecomRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Dtelecom.Contract.DtelecomCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Dtelecom *DtelecomRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dtelecom.Contract.DtelecomTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Dtelecom *DtelecomRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Dtelecom.Contract.DtelecomTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Dtelecom *DtelecomCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Dtelecom.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Dtelecom *DtelecomTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Dtelecom.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Dtelecom *DtelecomTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Dtelecom.Contract.contract.Transact(opts, method, params...)
}

// ClientByAddress is a free data retrieval call binding the contract method 0x9373e73d.
//
// Solidity: function clientByAddress(address ) view returns(uint256 limit, uint256 until, bool active, string key)
func (_Dtelecom *DtelecomCaller) ClientByAddress(opts *bind.CallOpts, arg0 common.Address) (struct {
	Limit  *big.Int
	Until  *big.Int
	Active bool
	Key    string
}, error) {
	var out []interface{}
	err := _Dtelecom.contract.Call(opts, &out, "clientByAddress", arg0)

	outstruct := new(struct {
		Limit  *big.Int
		Until  *big.Int
		Active bool
		Key    string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Limit = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Until = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Active = *abi.ConvertType(out[2], new(bool)).(*bool)
	outstruct.Key = *abi.ConvertType(out[3], new(string)).(*string)

	return *outstruct, err

}

// ClientByAddress is a free data retrieval call binding the contract method 0x9373e73d.
//
// Solidity: function clientByAddress(address ) view returns(uint256 limit, uint256 until, bool active, string key)
func (_Dtelecom *DtelecomSession) ClientByAddress(arg0 common.Address) (struct {
	Limit  *big.Int
	Until  *big.Int
	Active bool
	Key    string
}, error) {
	return _Dtelecom.Contract.ClientByAddress(&_Dtelecom.CallOpts, arg0)
}

// ClientByAddress is a free data retrieval call binding the contract method 0x9373e73d.
//
// Solidity: function clientByAddress(address ) view returns(uint256 limit, uint256 until, bool active, string key)
func (_Dtelecom *DtelecomCallerSession) ClientByAddress(arg0 common.Address) (struct {
	Limit  *big.Int
	Until  *big.Int
	Active bool
	Key    string
}, error) {
	return _Dtelecom.Contract.ClientByAddress(&_Dtelecom.CallOpts, arg0)
}

// GetAllNode is a free data retrieval call binding the contract method 0x63c393e0.
//
// Solidity: function getAllNode() view returns((uint256,bool,string)[])
func (_Dtelecom *DtelecomCaller) GetAllNode(opts *bind.CallOpts) ([]DtelecomNode, error) {
	var out []interface{}
	err := _Dtelecom.contract.Call(opts, &out, "getAllNode")

	if err != nil {
		return *new([]DtelecomNode), err
	}

	out0 := *abi.ConvertType(out[0], new([]DtelecomNode)).(*[]DtelecomNode)

	return out0, err

}

// GetAllNode is a free data retrieval call binding the contract method 0x63c393e0.
//
// Solidity: function getAllNode() view returns((uint256,bool,string)[])
func (_Dtelecom *DtelecomSession) GetAllNode() ([]DtelecomNode, error) {
	return _Dtelecom.Contract.GetAllNode(&_Dtelecom.CallOpts)
}

// GetAllNode is a free data retrieval call binding the contract method 0x63c393e0.
//
// Solidity: function getAllNode() view returns((uint256,bool,string)[])
func (_Dtelecom *DtelecomCallerSession) GetAllNode() ([]DtelecomNode, error) {
	return _Dtelecom.Contract.GetAllNode(&_Dtelecom.CallOpts)
}

// NodeByAddress is a free data retrieval call binding the contract method 0x3fd23654.
//
// Solidity: function nodeByAddress(address ) view returns(uint256 ip, bool active, string key)
func (_Dtelecom *DtelecomCaller) NodeByAddress(opts *bind.CallOpts, arg0 common.Address) (struct {
	Ip     *big.Int
	Active bool
	Key    string
}, error) {
	var out []interface{}
	err := _Dtelecom.contract.Call(opts, &out, "nodeByAddress", arg0)

	outstruct := new(struct {
		Ip     *big.Int
		Active bool
		Key    string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Ip = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Active = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.Key = *abi.ConvertType(out[2], new(string)).(*string)

	return *outstruct, err

}

// NodeByAddress is a free data retrieval call binding the contract method 0x3fd23654.
//
// Solidity: function nodeByAddress(address ) view returns(uint256 ip, bool active, string key)
func (_Dtelecom *DtelecomSession) NodeByAddress(arg0 common.Address) (struct {
	Ip     *big.Int
	Active bool
	Key    string
}, error) {
	return _Dtelecom.Contract.NodeByAddress(&_Dtelecom.CallOpts, arg0)
}

// NodeByAddress is a free data retrieval call binding the contract method 0x3fd23654.
//
// Solidity: function nodeByAddress(address ) view returns(uint256 ip, bool active, string key)
func (_Dtelecom *DtelecomCallerSession) NodeByAddress(arg0 common.Address) (struct {
	Ip     *big.Int
	Active bool
	Key    string
}, error) {
	return _Dtelecom.Contract.NodeByAddress(&_Dtelecom.CallOpts, arg0)
}

// AddClient is a paid mutator transaction binding the contract method 0x7d258e9b.
//
// Solidity: function addClient(address addr, uint256 limit, uint256 until, string key) returns()
func (_Dtelecom *DtelecomTransactor) AddClient(opts *bind.TransactOpts, addr common.Address, limit *big.Int, until *big.Int, key string) (*types.Transaction, error) {
	return _Dtelecom.contract.Transact(opts, "addClient", addr, limit, until, key)
}

// AddClient is a paid mutator transaction binding the contract method 0x7d258e9b.
//
// Solidity: function addClient(address addr, uint256 limit, uint256 until, string key) returns()
func (_Dtelecom *DtelecomSession) AddClient(addr common.Address, limit *big.Int, until *big.Int, key string) (*types.Transaction, error) {
	return _Dtelecom.Contract.AddClient(&_Dtelecom.TransactOpts, addr, limit, until, key)
}

// AddClient is a paid mutator transaction binding the contract method 0x7d258e9b.
//
// Solidity: function addClient(address addr, uint256 limit, uint256 until, string key) returns()
func (_Dtelecom *DtelecomTransactorSession) AddClient(addr common.Address, limit *big.Int, until *big.Int, key string) (*types.Transaction, error) {
	return _Dtelecom.Contract.AddClient(&_Dtelecom.TransactOpts, addr, limit, until, key)
}

// AddNode is a paid mutator transaction binding the contract method 0x0121daa1.
//
// Solidity: function addNode(uint256 ip, address addr, string key) returns()
func (_Dtelecom *DtelecomTransactor) AddNode(opts *bind.TransactOpts, ip *big.Int, addr common.Address, key string) (*types.Transaction, error) {
	return _Dtelecom.contract.Transact(opts, "addNode", ip, addr, key)
}

// AddNode is a paid mutator transaction binding the contract method 0x0121daa1.
//
// Solidity: function addNode(uint256 ip, address addr, string key) returns()
func (_Dtelecom *DtelecomSession) AddNode(ip *big.Int, addr common.Address, key string) (*types.Transaction, error) {
	return _Dtelecom.Contract.AddNode(&_Dtelecom.TransactOpts, ip, addr, key)
}

// AddNode is a paid mutator transaction binding the contract method 0x0121daa1.
//
// Solidity: function addNode(uint256 ip, address addr, string key) returns()
func (_Dtelecom *DtelecomTransactorSession) AddNode(ip *big.Int, addr common.Address, key string) (*types.Transaction, error) {
	return _Dtelecom.Contract.AddNode(&_Dtelecom.TransactOpts, ip, addr, key)
}

// RemoveClientByAddress is a paid mutator transaction binding the contract method 0x1e352258.
//
// Solidity: function removeClientByAddress(address addr) returns()
func (_Dtelecom *DtelecomTransactor) RemoveClientByAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Dtelecom.contract.Transact(opts, "removeClientByAddress", addr)
}

// RemoveClientByAddress is a paid mutator transaction binding the contract method 0x1e352258.
//
// Solidity: function removeClientByAddress(address addr) returns()
func (_Dtelecom *DtelecomSession) RemoveClientByAddress(addr common.Address) (*types.Transaction, error) {
	return _Dtelecom.Contract.RemoveClientByAddress(&_Dtelecom.TransactOpts, addr)
}

// RemoveClientByAddress is a paid mutator transaction binding the contract method 0x1e352258.
//
// Solidity: function removeClientByAddress(address addr) returns()
func (_Dtelecom *DtelecomTransactorSession) RemoveClientByAddress(addr common.Address) (*types.Transaction, error) {
	return _Dtelecom.Contract.RemoveClientByAddress(&_Dtelecom.TransactOpts, addr)
}

// RemoveNodeByAddress is a paid mutator transaction binding the contract method 0xc28a7e8f.
//
// Solidity: function removeNodeByAddress(address addr) returns()
func (_Dtelecom *DtelecomTransactor) RemoveNodeByAddress(opts *bind.TransactOpts, addr common.Address) (*types.Transaction, error) {
	return _Dtelecom.contract.Transact(opts, "removeNodeByAddress", addr)
}

// RemoveNodeByAddress is a paid mutator transaction binding the contract method 0xc28a7e8f.
//
// Solidity: function removeNodeByAddress(address addr) returns()
func (_Dtelecom *DtelecomSession) RemoveNodeByAddress(addr common.Address) (*types.Transaction, error) {
	return _Dtelecom.Contract.RemoveNodeByAddress(&_Dtelecom.TransactOpts, addr)
}

// RemoveNodeByAddress is a paid mutator transaction binding the contract method 0xc28a7e8f.
//
// Solidity: function removeNodeByAddress(address addr) returns()
func (_Dtelecom *DtelecomTransactorSession) RemoveNodeByAddress(addr common.Address) (*types.Transaction, error) {
	return _Dtelecom.Contract.RemoveNodeByAddress(&_Dtelecom.TransactOpts, addr)
}