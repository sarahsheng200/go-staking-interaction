package service

import (
	"fmt"
	"math/big"
	"staking-interaction/adapter"
	constant "staking-interaction/common"
	"staking-interaction/dto"
	"staking-interaction/model"
	"staking-interaction/repository"
)

type StakeService struct {
	contractInfo *adapter.ContractStakeInfo
}

func NewStakeService(
	contractInfo *adapter.ContractStakeInfo,
) *StakeService {
	return &StakeService{
		contractInfo: contractInfo,
	}
}

func (s *StakeService) Stake(amount int64, period uint8) (response *dto.StakeResponse, err error) {
	fmt.Println("---contract---", s.contractInfo)
	stakingContract := s.contractInfo.StakingContract
	auth := s.contractInfo.Auth

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
		FromAddress:     s.contractInfo.FromAddress,
		Method:          "stake",
	}

	return response, nil
}

func (s *StakeService) Withdraw(index big.Int) (response *dto.StakeResponse, err error) {
	stakingContract := s.contractInfo.StakingContract
	auth := s.contractInfo.Auth

	trans, err := stakingContract.Withdraw(auth, &index)

	if trans == nil || err != nil {
		return nil, fmt.Errorf("Withdraw transaction error: %w", err)
	}

	response = &dto.StakeResponse{
		Hash:            trans.Hash().String(),
		ContractAddress: constant.STAKE_CONTRACT_ADDRESS,
		FromAddress:     s.contractInfo.FromAddress,
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
