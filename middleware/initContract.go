package middleware

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"log"
	constant "staking-interaction/common"
	"staking-interaction/contracts"
)

func InitContract() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("InitContract-----")
		// 初始化客户端
		client, err := ethclient.Dial("https://data-seed-prebsc-1-s1.binance.org:8545")

		if err != nil {
			log.Fatalf("Failed to connect to the BSC network: %v", err)
		}

		// 加载私钥
		privateKey, err := crypto.HexToECDSA(constant.PRIVATE_KEY)
		if err != nil {
			log.Fatalf("Failed to parse private key: %v", err)
		}

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			log.Fatal("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		}

		fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
		log.Println("From address:", fromAddress.Hex())

		// 获取链ID
		chainID, err := client.ChainID(context.Background())
		if err != nil {
			log.Fatalf("Failed to get chain ID: %v", err)
		}

		// 创建授权事务
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			log.Fatalf("Failed to create authorized transactor: %v", err)
		}

		// 这里需要加载合约ABI和地址
		contractAddress := common.HexToAddress(constant.CONTRACT_ADDRESS)

		stakingContract, err := contracts.NewStaking(contractAddress, client)
		if err != nil {
			log.Fatalf("Failed to create staking contract: %v", err)
		}

		c.Set("stakingContract", stakingContract)
		c.Set("auth", auth)
		c.Set("fromAddress", fromAddress.String())

		fmt.Println("Go Ethereum SDK初始化完成")
	}

}
