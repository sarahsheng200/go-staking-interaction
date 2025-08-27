package adapter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	constant "staking-interaction/common"
	"staking-interaction/contracts/stake"
)

var (
	contractStakeInfo *ContractStakeInfo
)

type ContractStakeInfo struct {
	StakingContract *stake.Contracts
	Auth            *bind.TransactOpts
	FromAddress     common.Address
	Client          *ethclient.Client
}

func GetStakeContract() *ContractStakeInfo {
	return contractStakeInfo
}

func NewStakeContract() (*ContractStakeInfo, error) {
	clientInfo := GetInitClient()
	contractAddress := common.HexToAddress(constant.STAKE_CONTRACT_ADDRESS)
	//绑定合约实例
	//creates a new instance of Contracts, bound to a specific deployed contract
	stakingContract, err := stake.NewContracts(contractAddress, clientInfo.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create staking contract")
	}

	contractStakeInfo = &ContractStakeInfo{
		StakingContract: stakingContract,
		Auth:            clientInfo.auth,
		FromAddress:     clientInfo.fromAddress,
		Client:          clientInfo.client,
	}

	return contractStakeInfo, nil
}
