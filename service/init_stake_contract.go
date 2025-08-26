package service

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"staking-interaction/contracts/stake"
	"sync"
)

var (
	contractStakeInfo ContractStakeInfo
	dataStakeMutex    sync.RWMutex
)

type ContractStakeInfo struct {
	StakingContract *stake.Contracts
	Auth            *bind.TransactOpts
	FromAddress     string
	Client          *ethclient.Client
}

func GetStakeContract() ContractStakeInfo {
	return contractStakeInfo
}

func NewStakeContract(c ContractStakeInfo) {
	dataStakeMutex.Lock()
	defer dataStakeMutex.Unlock()

	contractStakeInfo = ContractStakeInfo{
		StakingContract: c.StakingContract,
		Auth:            c.Auth,
		FromAddress:     c.FromAddress,
		Client:          c.Client,
	}
}
