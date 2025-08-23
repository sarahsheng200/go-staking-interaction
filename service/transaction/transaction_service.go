package transaction

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	constant "staking-interaction/common"
	"staking-interaction/contracts/mtk"
	"staking-interaction/model"
	"time"
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
	toAddress := common.HexToAddress(req.ToAddress)
	amount := req.Amount

	clientInfo := model.GetInitClient()
	auth := clientInfo.Auth
	ethClient := clientInfo.Client

	contractAddr := common.HexToAddress(constant.TOKEN_CONTRACT_ADDRESS)
	mtkContract, err := mtk.NewContracts(contractAddr, ethClient)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "contract create failed", "error": err})
		return
	}
	tx, err := mtkContract.Transfer(auth, toAddress, amount)
	if tx == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "transfer failed", "error": err})
		return
	}
	sym, _ := mtkContract.Symbol(&bind.CallOpts{})
	decimal, _ := mtkContract.Decimals(&bind.CallOpts{})
	c.JSON(200, gin.H{"msg": "transfer successfully!", "data": ERCRes{Hash: tx.Hash().Hex(), Symbol: sym, Decimal: decimal}})

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
	toAddress := common.HexToAddress(req.ToAddress)
	amount := req.Amount

	clientInfo := model.GetInitClient()
	ethClient := clientInfo.Client
	fromAddress := common.HexToAddress(clientInfo.FromAddress)
	// 1. 准备交易参数
	// 1.1 动态获取nonce
	initialNonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	// 1.2 获取当前Gas价格（BSC网络）
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "获取Gas价格失败: " + err.Error()})
		return
	}
	// 1.3 BNB普通转账固定GasLimit为21000
	gasLimit := uint64(21000)
	// 2. 创建BNB转账交易（普通交易，不涉及合约）
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    initialNonce,
		To:       &toAddress,
		Value:    amount, // 转账金额（wei）
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     nil, // 普通转账无数据
	})
	// 3. 签名交易（使用BSC链ID）
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(clientInfo.ChainID), clientInfo.PrivateKey)
	if signedTx == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "signed Tx failed", "error": err})
		return
	}
	// 4. 发送交易到BSC网络
	if err := ethClient.SendTransaction(context.Background(), signedTx); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "send transaction failed", "error": err.Error()})
		return
	}
	success, err := checkTxStatus(ethClient, signedTx.Hash())
	if success {
		c.JSON(200, gin.H{"msg": "transfer successfully!", "data": signedTx.Hash().Hex()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"msg": "check transaction failed", "error": err})
	// 方案二：使用预签名生成hash的方式发送交易，转账之前提前生成了hash
	// 好处是：因为交易上链稳定是一个比较耗时的操作，大多数方式是发完交易区一直等待稳定，
	// 获取交易状态，一笔交易整体耗时可能是3~5s。假如现在有10000笔交易需要执行。
	// 需要有一个错误处理脚本去轮询失败的交易，重新发放，通过hash查询交易状态，如果失败，重新发放；getTransactionReceipt方法
}

func checkTxStatus(client *ethclient.Client, txHash common.Hash) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("transfer timeout")
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt.Status == types.ReceiptStatusSuccessful, nil
			}
			// 忽略"交易未找到"的错误（可能还在pending）
			if err != ethereum.NotFound {
				return false, err
			}
		}
	}
}
