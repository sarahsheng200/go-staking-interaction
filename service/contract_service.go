package service

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	constant "staking-interaction/common"
	"staking-interaction/contracts"
	"staking-interaction/model"
	"staking-interaction/repository"
	"strconv"
	"time"
)

func Stake(c *gin.Context) {
	var request model.StakeRequest
	var response model.Response
	stakingContract := c.MustGet("stakingContract").(*contracts.Contracts)
	auth := c.MustGet("auth").(*bind.TransactOpts)

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}

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
	response.ContractAddress = constant.CONTRACT_ADDRESS
	response.FromAddress = c.MustGet("fromAddress").(string)
	response.GasUsed = float64(trans.Gas())
	response.Method = "stake"

	StoreStakeInfo(response)

	c.JSON(http.StatusOK, gin.H{"msg": "Stake success!", "data": response})
}

func Withdraw(c *gin.Context) {
	var request model.WithDrawnRequest
	var response model.Response

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}
	indexStr := c.DefaultPostForm("index", "0")
	index, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "invalid index", "error": err.Error()})
		return
	}
	stakingContract := c.MustGet("stakingContract").(*contracts.Contracts)
	auth := c.MustGet("auth").(*bind.TransactOpts)

	trans, err := stakingContract.Withdraw(auth, big.NewInt(index))

	if trans == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "withdrawn transaction error", "error": err.Error()})
		return
	}

	response.Hash = trans.Hash().String()
	response.ContractAddress = constant.CONTRACT_ADDRESS
	response.FromAddress = c.MustGet("fromAddress").(string)
	response.GasUsed = float64(trans.Gas())
	response.Method = "withdraw"

	StoreStakeInfo(response)

	c.JSON(http.StatusOK, gin.H{"msg": "Withdrawn success!", "data": response})
}

func StoreStakeInfo(response model.Response) {
	var stake model.Stake
	stake.Hash = response.Hash
	stake.GasUsed = response.GasUsed
	stake.FromAddress = response.FromAddress
	stake.ContractAddress = response.ContractAddress
	stake.Method = response.Method
	stake.Timestamp = time.Now()
	repository.AddStakeInfo(stake)
}

func GetAllStakesById(c *gin.Context) {
	id := c.Param("id")
	stake := repository.GetAllStakesById(id)
	c.JSON(http.StatusOK, gin.H{"msg": "success", "data": stake})
}
