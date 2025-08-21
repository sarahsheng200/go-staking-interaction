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
	constant "staking-interaction/common"
	airdropModel "staking-interaction/model/airdrop"
	"sync"
	"time"
)

func GenerateMultiWallets(c *gin.Context) {
	var request airdropModel.Request
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
	var request airdropModel.Request
	var responses []airdropModel.Response
	var wg sync.WaitGroup
	var mu sync.Mutex // 保护response切片的并发写入
	var amounts []*big.Int

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"msg": "request body invalid", "err": err})
	}

	reqCount := request.Count
	reqBatchSize := request.BatchSize
	reqAmount := request.Amount

	if reqCount <= 0 || reqBatchSize <= 0 || reqAmount == nil || reqAmount.Cmp(big.NewInt(0)) <= 0 {
		c.AbortWithStatusJSON(500, gin.H{"msg": "request params invalid: count, batchSize must be positive and amount must be greater than 0"})
	}

	walletAddresses, err := getMultiWallets(reqCount)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"msg": "generate wallet failed", "err": err})
	}

	for range walletAddresses {
		amounts = append(amounts, reqAmount)
	}

	ctx, cancel := context.WithTimeout(c, 10*time.Second) // 设置整体超时
	defer cancel()

	batchNum := len(walletAddresses) / reqBatchSize
	startIndex := 0
	endIndex := reqBatchSize
	contract := airdropModel.GetInitContract()
	fromAddr := common.HexToAddress(contract.FromAddress)

	initialNonce, err := contract.Client.PendingNonceAt(ctx, fromAddr)

	for i := 0; i < batchNum+1 && startIndex <= endIndex; i++ {
		var batchAddress []common.Address
		var batchAmounts []*big.Int
		if startIndex < endIndex {
			batchAddress = walletAddresses[startIndex:endIndex]
			batchAmounts = amounts[startIndex:endIndex]
		} else {
			batchAddress = walletAddresses[startIndex-1 : endIndex]
			batchAmounts = amounts[startIndex-1 : endIndex]
		}

		batchAuth := *contract.Auth
		batchNonce := initialNonce + uint64(i)
		batchAuth.Nonce = big.NewInt(int64(batchNonce))

		wg.Add(1)
		go func(idx int, addrs []common.Address, amts []*big.Int, auth *bind.TransactOpts) {
			defer wg.Done()
			res := processAirdropERC20(idx, addrs, amts, contract, auth)
			// 线程安全地收集结果
			mu.Lock()
			responses = append(responses, res)
			mu.Unlock()
		}(i, batchAddress, batchAmounts, &batchAuth)
		startIndex = endIndex
		endIndex = startIndex + reqBatchSize
		if endIndex > len(walletAddresses) {
			endIndex = len(walletAddresses)
		}

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
		c.JSON(200, gin.H{"msg": "success", "data": responses})
	case <-ctx.Done():
		c.JSON(504, gin.H{"msg": "airdrop timeout", "err": ctx.Err().Error()})
		return
	case <-time.After(10 * time.Second):
		c.JSON(504, gin.H{"msg": "response timeout"})
		return
	}

}

func processAirdropERC20(idx int, batchAddress []common.Address, batchAmounts []*big.Int, contract airdropModel.ContractInitInfo, auth *bind.TransactOpts) (response airdropModel.Response) {
	airdropContract := contract.AirdropContract
	trans, err := airdropContract.AirdropERC20(auth, batchAddress, batchAmounts)

	if trans == nil || err != nil {
		return airdropModel.Response{
			BatchNum:        idx,
			Error:           fmt.Sprintf("airdrop failed: ", err),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
		}
	} else {
		return airdropModel.Response{
			BatchNum:        idx,
			Hash:            trans.Hash().Hex(),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
		}
	}

}
