package service

import (
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	constant "staking-interaction/common"
	"staking-interaction/model"
	"staking-interaction/repository"
)

func Stake(c *gin.Context) {
	var request model.StakeRequest
	var response model.Response

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}

	contract := model.GetInitContract()
	stakingContract := contract.StakingContract
	auth := contract.Auth

	trans, err := stakingContract.Stake(
		auth,
		big.NewInt(request.Amount),
		request.Period,
	)

	if trans == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "stake transaction error", "error": err.Error()})
		return
	}

	response.Hash = trans.Hash().String()
	response.ContractAddress = constant.STAKE_CONTRACT_ADDRESS
	response.FromAddress = contract.FromAddress
	response.Method = "stake"

	c.JSON(http.StatusOK, gin.H{"msg": "Stake success!", "data": response})
}

func Withdraw(c *gin.Context) {
	var request model.WithDrawnRequest
	var response model.Response

	contract := model.GetInitContract()
	stakingContract := contract.StakingContract
	auth := contract.Auth

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}

	trans, err := stakingContract.Withdraw(auth, &request.Index)

	if trans == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "withdrawn transaction error", "error": err.Error()})
		return
	}

	response.Hash = trans.Hash().String()
	response.ContractAddress = constant.STAKE_CONTRACT_ADDRESS
	response.FromAddress = contract.FromAddress
	response.Method = "withdraw"

	c.JSON(http.StatusOK, gin.H{"msg": "Withdrawn success!", "data": response})
}

func StoreStakeInfo(stake model.Stake) {
	repository.AddStakeInfo(stake)
}

func GetAllStakesByFromAddress(c *gin.Context) {
	id := c.Param("id")
	stake := repository.GetAllStakesByFromAddress(id)
	c.JSON(http.StatusOK, gin.H{"msg": "success", "data": stake})
}
