package airdrop

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"staking-interaction/contracts/airdrop"
	"sync"
)

var (
	contractInfo ContractInitInfo
	dataMutex    sync.RWMutex
	nouce        int64
)

type ContractInitInfo struct {
	AirdropContract *airdrop.Contracts
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

func GetNouce() int64 {
	return nouce + 1
}

func InitNouce() {
	nouce = 0
}
