package service

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	contract "staking-interaction/contracts/airdrop"
	"sync"
)

var (
	contractAirdropInfo AirdropContractInfo
	dataAirdropMutex    sync.RWMutex
)

type AirdropContractInfo struct {
	AirdropContract *contract.Contracts
	Auth            *bind.TransactOpts
	FromAddress     string
	Client          *ethclient.Client
}

func GetAirdropContract() AirdropContractInfo {
	return contractAirdropInfo
}

func NewAirdropContract(c AirdropContractInfo) {
	dataAirdropMutex.Lock()
	defer dataAirdropMutex.Unlock()

	contractAirdropInfo = AirdropContractInfo{
		AirdropContract: c.AirdropContract,
		Auth:            c.Auth,
		FromAddress:     c.FromAddress,
		Client:          c.Client,
	}
}
