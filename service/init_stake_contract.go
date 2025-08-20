package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	constant "staking-interaction/common"
	"staking-interaction/contracts/mtk"
	"staking-interaction/model"
)

func InitStakeContract() {
	log.Println("InitContract-----")
	// 初始化客户端
	client, err := ethclient.Dial(constant.RAW_URL)
	if err != nil {
		log.Fatalf("Failed to connect to the BSC network: %v", err)
	}
	defer client.Close()

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

	// 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	// create a transaction signer from a single private key.
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	contractAddress := common.HexToAddress(constant.STAKE_CONTRACT_ADDRESS)
	//绑定合约实例
	//creates a new instance of Contracts, bound to a specific deployed contract
	stakingContract, err := mtk.NewContracts(contractAddress, client)
	if err != nil {
		log.Fatalf("Failed to create staking contract: %v", err)
	}

	model.NewInitContract(model.ContractInitInfo{
		StakingContract: stakingContract,
		Auth:            auth,
		FromAddress:     fromAddress.String(),
		Client:          client,
	})

	fmt.Println("Go Ethereum SDK初始化完成")
}
