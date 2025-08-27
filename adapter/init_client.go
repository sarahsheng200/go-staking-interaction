package adapter

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	constant "staking-interaction/common"
)

type InitClient struct {
	auth        *bind.TransactOpts
	fromAddress common.Address
	client      *ethclient.Client
	privateKey  *ecdsa.PrivateKey
	chainID     *big.Int
}

var (
	clientInfo *InitClient
)

func NewInitClient() (*InitClient, error) {
	log.Println("NewInitClient-----")
	// 初始化客户端
	ethClient, err := ethclient.Dial(constant.RAW_URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the stake contract: %v", err)
	}

	// 加载私钥
	privateKey, err := crypto.HexToECDSA(constant.PRIVATE_KEY)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取链ID
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// create a transaction signer from a single private key.
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorized transactor: %v", err)
	}

	clientInfo = &InitClient{
		auth:        auth,
		client:      ethClient,
		fromAddress: fromAddress,
		privateKey:  privateKey,
		chainID:     chainID,
	}

	return clientInfo, nil
}
func GetInitClient() *InitClient {
	return clientInfo
}

func (c *InitClient) CloseInitClient() {
	c.client.Close()
}
