package airdrop

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	contract "staking-interaction/contracts/airdrop"
	"sync"
)

var (
	contractInfo ContractInitInfo
	dataMutex    sync.RWMutex
)

type ContractInitInfo struct {
	AirdropContract *contract.Contracts
	Auth            *bind.TransactOpts
	FromAddress     string
	Client          *ethclient.Client
}

func GetInitContract() ContractInitInfo {
	return contractInfo
}

func NewInitContract(c ContractInitInfo) {
	dataMutex.Lock()
	defer dataMutex.Unlock()

	contractInfo = ContractInitInfo{
		AirdropContract: c.AirdropContract,
		Auth:            c.Auth,
		FromAddress:     c.FromAddress,
		Client:          c.Client,
	}
}
