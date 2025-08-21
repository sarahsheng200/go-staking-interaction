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
	"staking-interaction/contracts/airdrop"
	"staking-interaction/contracts/stake"
	airdropModel "staking-interaction/model/airdrop"
	stakeModel "staking-interaction/model/stake"
)

func InitContracts() *ethclient.Client {
	log.Println("InitStakeContract-----")
	// 初始化客户端
	ethClient, err := ethclient.Dial(constant.RAW_URL)
	if err != nil {
		log.Fatalf("Failed to connect to the stake contract: %v", err)
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

	// 获取链ID
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	// create a transaction signer from a single private key.
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	createStakeContract(auth, fromAddress, ethClient)
	createAirdropContract(auth, fromAddress, ethClient)

	return ethClient
}

func createStakeContract(auth *bind.TransactOpts, fromAddress common.Address, client *ethclient.Client) {
	contractAddress := common.HexToAddress(constant.STAKE_CONTRACT_ADDRESS)
	//绑定合约实例
	//creates a new instance of Contracts, bound to a specific deployed contract
	stakingContract, err := stake.NewContracts(contractAddress, client)
	if err != nil {
		log.Fatalf("Failed to create staking contract: %v", err)
	}

	stakeModel.NewInitContract(stakeModel.ContractInitInfo{
		StakingContract: stakingContract,
		Auth:            auth,
		FromAddress:     fromAddress.String(),
		Client:          client,
	})

	fmt.Println("create stake contract successfully!")
}

func createAirdropContract(auth *bind.TransactOpts, fromAddress common.Address, client *ethclient.Client) {
	contractAddress := common.HexToAddress(constant.AIRDROP_CONTRACT_ADDRESS)
	airdropContract, err := airdrop.NewContracts(contractAddress, client)
	if err != nil {
		log.Fatalf("Failed to create airdrop contract: %v", err)
	}

	airdropModel.NewInitContract(airdropModel.ContractInitInfo{
		AirdropContract: airdropContract,
		Auth:            auth,
		FromAddress:     fromAddress.String(),
		Client:          client,
	})

	fmt.Println("create airdrop contract successfully!")
}
