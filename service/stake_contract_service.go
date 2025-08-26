package service

import (
	"fmt"
	"math/big"
	constant "staking-interaction/common"
	"staking-interaction/dto"
	"staking-interaction/model"
	"staking-interaction/repository"
)

func Stake(amount int64, period uint8) (response *dto.StakeResponse, err error) {
	contract := GetStakeContract()
	fmt.Println("---contract---", contract)
	stakingContract := contract.StakingContract
	auth := contract.Auth

	trans, err := stakingContract.Stake(
		auth,
		big.NewInt(amount),
		period,
	)

	if trans == nil || err != nil {
		return nil, fmt.Errorf("Stake transaction error: %w", err)
	}

	response = &dto.StakeResponse{
		Hash:            trans.Hash().String(),
		ContractAddress: constant.STAKE_CONTRACT_ADDRESS,
		FromAddress:     contract.FromAddress,
		Method:          "stake",
	}

	return response, nil
}

func Withdraw(index big.Int) (response *dto.StakeResponse, err error) {
	contract := GetStakeContract()
	stakingContract := contract.StakingContract
	auth := contract.Auth

	trans, err := stakingContract.Withdraw(auth, &index)

	if trans == nil || err != nil {
		return nil, fmt.Errorf("Withdraw transaction error: %w", err)
	}

	response = &dto.StakeResponse{
		Hash:            trans.Hash().String(),
		ContractAddress: constant.STAKE_CONTRACT_ADDRESS,
		FromAddress:     contract.FromAddress,
		Method:          "withdraw",
	}

	return response, nil
}

func StoreStakeInfo(stake model.Stake) {
	repository.AddStakeInfo(stake)
}

func GetAllStakesByFromAddress(id string) (response model.Stake) {
	stake := repository.GetAllStakesByFromAddress(id)
	return stake
}
