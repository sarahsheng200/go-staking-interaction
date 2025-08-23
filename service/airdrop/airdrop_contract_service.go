package airdrop

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	constant "staking-interaction/common"
	model "staking-interaction/model/airdrop"
	"sync"
	"time"
)

func GenerateMultiWallets(c *gin.Context) {
	var request model.Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(500, gin.H{"msg": "request body invalid", "err": err})
	}
	wallets, err := getMultiWallets(request.Count)
	if err != nil {
		c.JSON(500, gin.H{"msg": "generate wallet failed", "err": err})
	}
	c.JSON(200, gin.H{"msg": "generate success!", "data": wallets})

}

func getMultiWallets(count int) (walletAddresses []common.Address, err error) {
	for i := 0; i < count; i++ {
		// 1. 生成随机私钥（secp256k1曲线）
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}

		// 2. 从私钥推导公钥（未压缩格式）
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, err
		}

		// 3. 从公钥推导以太坊地址（0x前缀）
		address := crypto.PubkeyToAddress(*publicKeyECDSA)

		walletAddresses = append(walletAddresses, address)
	}
	return walletAddresses, nil
}

func AirdropERC20(c *gin.Context) {
	var (
		request   model.Request
		responses []model.Response
		wg        sync.WaitGroup
		mu        sync.Mutex // 保护response切片的并发写入
		amounts   []*big.Int
	)

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"msg": "request body invalid", "err": err})
	}

	reqCount := request.Count
	reqBatchSize := request.BatchSize
	reqAmount := request.Amount

	if request.Count <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "count 必须大于 0"})
		return
	}
	if request.BatchSize <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "batchSize 必须大于 0"})
		return
	}
	if request.Amount == nil || request.Amount.Cmp(big.NewInt(0)) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "amount 必须大于 0"})
		return
	}

	// generate multiple wallets
	walletAddresses, err := getMultiWallets(reqCount)
	if err != nil || walletAddresses == nil || len(walletAddresses) == 0 {
		c.AbortWithStatusJSON(500, gin.H{"msg": "generate wallet failed", "err": err})
		return
	}
	// generate related accounts
	for range walletAddresses {
		amounts = append(amounts, reqAmount)
	}
	addressLen := len(walletAddresses)
	contract := model.GetInitContract()
	fromAddr := common.HexToAddress(contract.FromAddress)

	ctx, cancel := context.WithTimeout(c, 10*time.Second) // 设置整体超时
	defer cancel()

	initialNonce, err := contract.Client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "获取 nonce 失败", "err": err.Error()})
		return
	}

	batchNum := (addressLen + reqBatchSize - 1) / reqBatchSize
	startIndex := 0
	currentNonce := initialNonce
	for i := 0; i < batchNum && startIndex < addressLen; i++ {
		endIndex := startIndex + reqBatchSize
		if endIndex > addressLen {
			endIndex = addressLen
		}

		batchAddress := walletAddresses[startIndex:endIndex]
		batchAmounts := amounts[startIndex:endIndex]
		batchNonce := currentNonce
		currentNonce++

		wg.Add(1)
		go func(idx int, addrs []common.Address, amts []*big.Int, nonce uint64) {
			defer wg.Done()
			batchAuth := *contract.Auth
			batchAuth.Nonce = big.NewInt(int64(nonce))
			batchAuth.Context = ctx

			res := processAirdropERC20(idx, addrs, amts, contract, &batchAuth)
			// 线程安全地收集结果
			mu.Lock()
			responses = append(responses, res)
			mu.Unlock()
		}(i, batchAddress, batchAmounts, batchNonce)

		startIndex = endIndex
	}

	done := make(chan struct{})
	// 等待所有批次完成并返回结果
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待结果或超时
	select {
	case <-done:
		successCount := 0
		failedCount := 0
		for _, res := range responses {
			if res.Error == "" {
				successCount++
			} else {
				failedCount++
			}
		}
		c.JSON(200, gin.H{
			"msg":              "airdrop success!",
			"completedBatches": len(responses),
			"successBatches":   successCount,
			"failBatches":      failedCount,
			"data":             responses,
		})
	case <-ctx.Done():
		c.JSON(504, gin.H{
			"msg":              "airdrop timeout",
			"completedBatches": len(responses),
			"err":              ctx.Err().Error(),
			"data":             responses,
		})
	}

}

func processAirdropERC20(idx int, batchAddress []common.Address, batchAmounts []*big.Int, contract model.ContractInitInfo, auth *bind.TransactOpts) (response model.Response) {
	airdropContract := contract.AirdropContract
	trans, err := airdropContract.AirdropERC20(auth, batchAddress, batchAmounts)
	if trans == nil || err != nil {
		return model.Response{
			BatchNum:        idx,
			Error:           fmt.Sprintf("airdrop failed: %v", err),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
		}
	} else {
		return model.Response{
			BatchNum:        idx,
			Hash:            trans.Hash().Hex(),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
			Error:           "",
		}
	}

}
