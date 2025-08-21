package service

import (
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	constant "staking-interaction/common"
	stakeModel "staking-interaction/model/stake"
	"staking-interaction/repository"
)

func Stake(c *gin.Context) {
	var request stakeModel.StakeRequest
	var response stakeModel.Response

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}

	contract := stakeModel.GetInitContract()
	stakingContract := contract.StakingContract
	auth := contract.Auth

	trans, err := stakingContract.Stake(
		auth,
		big.NewInt(request.Amount),
		request.Period,
	)

	if trans == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "stake transaction error", "error": err})
		return
	}

	response.Hash = trans.Hash().String()
	response.ContractAddress = constant.STAKE_CONTRACT_ADDRESS
	response.FromAddress = contract.FromAddress
	response.Method = "stake"

	c.JSON(http.StatusOK, gin.H{"msg": "Stake success!", "data": response})
}

func Withdraw(c *gin.Context) {
	var request stakeModel.WithDrawnRequest
	var response stakeModel.Response

	contract := stakeModel.GetInitContract()
	stakingContract := contract.StakingContract
	auth := contract.Auth

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}

	trans, err := stakingContract.Withdraw(auth, &request.Index)

	if trans == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "withdrawn transaction error", "error": err})
		return
	}

	response.Hash = trans.Hash().String()
	response.ContractAddress = constant.STAKE_CONTRACT_ADDRESS
	response.FromAddress = contract.FromAddress
	response.Method = "withdraw"

	c.JSON(http.StatusOK, gin.H{"msg": "Withdrawn success!", "data": response})
}

func StoreStakeInfo(stake stakeModel.Stake) {
	repository.AddStakeInfo(stake)
}

func GetAllStakesByFromAddress(c *gin.Context) {
	id := c.Param("id")
	stake := repository.GetAllStakesByFromAddress(id)
	c.JSON(http.StatusOK, gin.H{"msg": "success", "data": stake})
}
