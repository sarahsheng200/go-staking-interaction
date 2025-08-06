package service

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	constant "staking-interaction/common"
	"staking-interaction/contracts"
	"strconv"
)

type Response struct {
	Hash            string `json:"hash"`
	ContractAddress string `json:"contractAddress"`
	FromAddress     string `json:"fromAddress"`
	GasUsed         uint64 `json:"gasUsed"`
}
type StakeRequest struct {
	Amount int64 `json:"amount"`
	Period uint8 `json:"period"`
}

type WithDrawnRequest struct {
	Index int64 `json:"index"`
}

func Stake(c *gin.Context) {
	var request StakeRequest
	var response Response
	stakingContract := c.MustGet("stakingContract").(*contracts.Staking)
	auth := c.MustGet("auth").(*bind.TransactOpts)

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "request body invalid"})
	}

	trans, err := stakingContract.Stake(
		auth,
		big.NewInt(request.Amount),
		request.Period,
	)
	if trans == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"stake transaction error": err.Error()})

	}

	response.Hash = trans.Hash().String()
	response.ContractAddress = constant.CONTRACT_ADDRESS
	response.FromAddress = c.MustGet("fromAddress").(string)
	response.GasUsed = trans.Gas()

	c.JSON(http.StatusOK, gin.H{"msg": "Stake success!", "data": response})
}

func Withdraw(c *gin.Context) {
	var request WithDrawnRequest
	var response Response

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "request body invalid"})
	}
	indexStr := c.DefaultPostForm("index", "0")
	index, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
	}
	stakingContract := c.MustGet("stakingContract").(*contracts.Staking)
	auth := c.MustGet("auth").(*bind.TransactOpts)

	trans, err := stakingContract.Withdraw(auth, big.NewInt(index))

	if trans == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"withdrawn transaction error": err.Error()})
	}

	response.Hash = trans.Hash().String()
	response.ContractAddress = constant.CONTRACT_ADDRESS
	response.FromAddress = c.MustGet("fromAddress").(string)
	response.GasUsed = trans.Gas()
}
