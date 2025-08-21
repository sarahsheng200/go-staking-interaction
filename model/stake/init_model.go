package stake

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"staking-interaction/contracts/stake"
	"sync"
)

var (
	contractInfo ContractInitInfo
	dataMutex    sync.RWMutex
)

type ContractInitInfo struct {
	StakingContract *stake.Contracts
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
		StakingContract: c.StakingContract,
		Auth:            c.Auth,
		FromAddress:     c.FromAddress,
		Client:          c.Client,
	}
}
