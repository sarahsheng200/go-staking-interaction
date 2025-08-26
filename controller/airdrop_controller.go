package controller

import (
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"staking-interaction/dto"
	"staking-interaction/service"
)

func GenerateMultiWallets(c *gin.Context) {
	var request dto.AirdropRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(500, gin.H{"msg": "request body invalid", "err": err})
	}
	wallets, err := service.GetMultiWallets(request.Count)
	if err != nil {
		c.JSON(500, gin.H{"msg": "generate wallet failed", "err": err})
	}
	c.JSON(200, gin.H{"msg": "generate success!", "data": wallets})

}

func AirdropERC20(c *gin.Context) {
	var request dto.AirdropRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"msg": "request body invalid", "err": err})
	}

	reqCount := request.Count
	reqBatchSize := request.BatchSize
	reqAmount := request.Amount

	if reqCount <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "count 必须大于 0"})
		return
	}
	if reqBatchSize <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "batchSize 必须大于 0"})
		return
	}
	if reqAmount == nil || reqAmount.Cmp(big.NewInt(0)) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "amount 必须大于 0"})
		return
	}

	responses, err := service.AirdropERC20(reqCount, reqBatchSize, reqAmount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Airdrop failed", "err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, responses)

}

func AirdropBNB(c *gin.Context) {
	var (
		request dto.AirdropRequest
	)

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"msg": "request body invalid", "err": err})
	}

	reqCount := request.Count
	reqBatchSize := request.BatchSize
	reqAmount := request.Amount

	if reqCount <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "count 必须大于 0"})
		return
	}
	if reqBatchSize <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "batchSize 必须大于 0"})
		return
	}
	if reqAmount == nil || reqAmount.Cmp(big.NewInt(0)) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "amount 必须大于 0"})
		return
	}

	responses, err := service.AirdropBNB(reqCount, reqBatchSize, reqAmount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Airdrop failed", "err": err})
		return
	}
	c.JSON(http.StatusOK, responses)

}
