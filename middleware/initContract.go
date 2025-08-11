package middleware

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"log"
	"math/big"
	constant "staking-interaction/common"
	"staking-interaction/contracts"
	"strings"
	"time"
)

func InitContract() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("InitContract-----")
		// 初始化客户端

		client, err := ethclient.Dial("https://data-seed-prebsc-2-s1.binance.org:8545")
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

		// 创建授权事务
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
		if err != nil {
			log.Fatalf("Failed to create authorized transactor: %v", err)
		}

		// 这里需要加载合约ABI和地址
		contractAddress := common.HexToAddress(constant.CONTRACT_ADDRESS)

		stakingContract, err := contracts.NewContracts(contractAddress, client)
		if err != nil {
			log.Fatalf("Failed to create staking contract: %v", err)
		}

		c.Set("stakingContract", stakingContract)
		c.Set("auth", auth)
		c.Set("fromAddress", fromAddress.String())
		c.Set("client", client)

		fmt.Println("Go Ethereum SDK初始化完成")
	}
}

func ListenToEvents() {
	client, err := ethclient.Dial("https://bsc-testnet-rpc.publicnode.com")
	if err != nil {
		log.Fatalf("ListenToStakedEvent: Failed to connect to the BSC network: %v", err)
	}
	defer client.Close()

	contractAddress := common.HexToAddress(constant.CONTRACT_ADDRESS)

	contractABI, err := abi.JSON(strings.NewReader(contracts.ContractsMetaData.ABI))
	if err != nil {
		log.Fatalf("ListenToStakedEvent: Failed to parse contract ABI: %v", err)
	}

	stakedEventName := "Staked"
	stakedEventId := contractABI.Events[stakedEventName].ID

	withdrawnEventName := "Withdrawn"
	withdrawnEventId := contractABI.Events[withdrawnEventName].ID

	startBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("ListenToStakedEvent: Failed to get block number: %v", err)
	}
	fmt.Println("start block number:", startBlock)

	for {
		currentBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("ListenToStakedEvent: Failed to get block number, will retry after 10 sec: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if currentBlock > startBlock {
			for blockNum := startBlock + 1; blockNum <= currentBlock; blockNum++ {
				logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(blockNum)),
					ToBlock:   big.NewInt(int64(blockNum)),
					Addresses: []common.Address{contractAddress},
					Topics:    [][]common.Hash{{stakedEventId, withdrawnEventId}},
				})
				if err != nil {
					log.Fatalf("ListenToStakedEvent:  Failed to filter logs: BlockNumber is: %d, err is: %v", blockNum, err)
				}

				for _, l := range logs {
					switch l.Topics[0] {
					case stakedEventId:
						var event contracts.ContractsStaked
						if err := contractABI.UnpackIntoInterface(&event, stakedEventName, l.Data); err != nil {
							log.Printf("ListenToStakedEvent: Staked failed to unpack event: %v", err)
							continue
						}

						if len(l.Topics) >= 2 {
							event.User = common.HexToAddress(l.Topics[1].Hex())
						} else {
							log.Printf("ListenToStakedEvent: Staked failed to unpack event: %v", err)
							continue
						}

						fmt.Printf("Staked: block number: %d, transaction hash:%v, user:%v, amount:%s, timestamp=%d,  StakedIndex=%s\n",
							l.BlockNumber,
							l.TxHash.Hex(),
							event.User,
							event.Amount.String(),
							event.Timestamp.Int64(),
							event.StakeIndex.String(),
						)

					case withdrawnEventId:
						var event contracts.ContractsWithdrawn
						if err := contractABI.UnpackIntoInterface(&event, withdrawnEventName, l.Data); err != nil {
							log.Printf("ListenToStakedEvent: Withdrawn failed to unpack event: %v", err)
							continue
						}

						if len(l.Topics) >= 2 {
							event.User = common.HexToAddress(l.Topics[1].Hex())
						} else {
							log.Printf("ListenToStakedEvent: Withdrawn failed to unpack event: %v", err)
							continue
						}

						fmt.Printf("Withdrawn: block number: %d, transaction hash:%v, user:%v, amount:%s, StakedIndex=%s\n",
							l.BlockNumber,
							l.TxHash.Hex(),
							event.User,
							event.TotalAmount.String(),
							event.StakeIndex.String(),
						)

					}

				}
			}
			startBlock = currentBlock
		} else {
			time.Sleep(10 * time.Second)
		}

	}
}
