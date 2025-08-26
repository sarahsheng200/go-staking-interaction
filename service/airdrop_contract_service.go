package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	constant "staking-interaction/common"
	"staking-interaction/dto"
	"sync"
	"time"
)

func GetMultiWallets(count int) (walletAddresses []common.Address, err error) {
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

func AirdropERC20(reqCount int, reqBatchSize int, reqAmount *big.Int) (data *dto.AirdropResponse, err error) {
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex // 保护response切片的并发写入
		amounts   []*big.Int
		responses []dto.AirdropInfo
	)

	// generate multiple wallets
	walletAddresses, err := GetMultiWallets(reqCount)
	if err != nil || walletAddresses == nil || len(walletAddresses) == 0 {
		return nil, fmt.Errorf("generate wallet failed: %v", err)
	}

	// generate related accounts
	for range walletAddresses {
		amounts = append(amounts, reqAmount)
	}
	addressLen := len(walletAddresses)
	contract := GetAirdropContract()
	fromAddr := common.HexToAddress(contract.FromAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置整体超时
	defer cancel()

	initialNonce, err := contract.Client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("Retrieve nonce failed: %v", err)
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
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop success!",
			CompletedBatches: len(responses),
			SuccessBatches:   successCount,
			FailBatches:      failedCount,
			Data:             responses}
		return &resInfo, nil
	case <-ctx.Done():
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop timeout",
			CompletedBatches: len(responses),
			Data:             responses,
			Error:            ctx.Err().Error(),
		}
		return &resInfo, nil
	}
}

func processAirdropERC20(idx int, batchAddress []common.Address, batchAmounts []*big.Int, contract AirdropContractInfo, auth *bind.TransactOpts) (response dto.AirdropInfo) {
	airdropContract := contract.AirdropContract
	trans, err := airdropContract.AirdropERC20(auth, batchAddress, batchAmounts)
	if trans == nil || err != nil {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Error:           fmt.Sprintf("airdrop failed: %v", err),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
		}
	} else {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Hash:            trans.Hash().Hex(),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
			Error:           "",
		}
	}

}

func AirdropBNB(reqCount int, reqBatchSize int, reqAmount *big.Int) (data *dto.AirdropResponse, err error) {
	var (
		responses []dto.AirdropInfo
		wg        sync.WaitGroup
		mu        sync.Mutex // 保护response切片的并发写入
		amounts   []*big.Int
	)

	// generate multiple wallets
	walletAddresses, err := GetMultiWallets(reqCount)
	if err != nil || walletAddresses == nil || len(walletAddresses) == 0 {
		return nil, fmt.Errorf("generate wallet failed: %v", err)
	}
	// generate related accounts
	for range walletAddresses {
		amounts = append(amounts, reqAmount)
	}
	addressLen := len(walletAddresses)
	contract := GetAirdropContract()
	fromAddr := common.HexToAddress(contract.FromAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置整体超时
	defer cancel()

	initialNonce, err := contract.Client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("Retrieve nonce failed: %v", err)
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
			val := new(big.Int)
			val.Mul(reqAmount, big.NewInt(int64(len(batchAmounts))))
			batchAuth.Value = val

			res := processAirdropBNB(idx, addrs, amts, contract, &batchAuth)
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
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop success!",
			CompletedBatches: len(responses),
			SuccessBatches:   successCount,
			FailBatches:      failedCount,
			Data:             responses}
		return &resInfo, nil
	case <-ctx.Done():
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop timeout",
			CompletedBatches: len(responses),
			Data:             responses,
			Error:            ctx.Err().Error(),
		}
		return &resInfo, nil
	}

}

func processAirdropBNB(idx int, batchAddress []common.Address, batchAmounts []*big.Int, contract AirdropContractInfo, auth *bind.TransactOpts) (response dto.AirdropInfo) {
	airdropContract := contract.AirdropContract
	//trans, err := airdropContract.AirdropERC20(auth, batchAddress, batchAmounts)
	trans, err := airdropContract.AirdropBNB(auth, batchAddress, batchAmounts)
	if trans == nil || err != nil {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Error:           fmt.Sprintf("airdrop failed: %v", err),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
		}
	} else {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Hash:            trans.Hash().Hex(),
			ContractAddress: constant.AIRDROP_CONTRACT_ADDRESS,
			FromAddress:     contract.FromAddress,
			WalletAddress:   batchAddress,
			Error:           "",
		}
	}

}
