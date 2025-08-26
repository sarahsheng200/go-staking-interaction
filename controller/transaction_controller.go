package controller

import (
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"staking-interaction/service"
)

type ERCRequest struct {
	ToAddress string   `json:"toAddress"`
	Amount    *big.Int `json:"amount"`
}
type ERCRes struct {
	Hash    string `json:"hash"`
	Symbol  string `json:"symbol"`
	Decimal uint8  `json:"decimal"`
}

func SendErc20(c *gin.Context) {
	// 方案一：普通交易，与合约没关系，需要转账之后，等待交易是成功还是失败,
	// tx.wait();
	// 1. 实现转账，获取转账的hash（交易完成后才有hash）
	// 2. 通过hash查询交易是否成功
	// 创建转账交易
	var req ERCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := service.SendErc20(req.ToAddress, req.Amount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "Airdrop failed", "err": err})
		return
	}

	c.JSON(http.StatusOK, res)
}

func SendBNB(c *gin.Context) {
	// 方案一：普通交易，与合约没关系，需要转账之后，等待交易是成功还是失败,
	// tx.wait();
	// 1. 实现转账，获取转账的hash（交易完成后才有hash）
	// 2. 通过hash查询交易是否成功
	var req ERCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := service.SendBNB(req.ToAddress, req.Amount)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "Airdrop failed", "err": err})
		return
	}

	c.JSON(http.StatusOK, res)
}
