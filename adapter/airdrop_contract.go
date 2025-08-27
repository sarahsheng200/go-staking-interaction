package adapter

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	constant "staking-interaction/common"
	"staking-interaction/contracts/airdrop"
	"sync"
)

var (
	airdropContractInfo *AirdropContractInfo
)

type AirdropContractInfo struct {
	airdropContract *airdrop.Contracts
	auth            *bind.TransactOpts
	fromAddress     common.Address
	client          *ethclient.Client
	mu              sync.Mutex
}

func NewAirdropContract() (*AirdropContractInfo, error) {
	clientInfo := GetInitClient()
	contractAddress := common.HexToAddress(constant.AIRDROP_CONTRACT_ADDRESS)
	airdropContract, err := airdrop.NewContracts(contractAddress, clientInfo.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create airdrop contract")
	}

	if airdropContract == nil {
		return nil, fmt.Errorf("airdropContract should not be nil")
	}
	if clientInfo.auth == nil {
		return nil, fmt.Errorf("auth should not be nil")
	}
	if clientInfo.client == nil {
		return nil, fmt.Errorf("client should not be nil")
	}
	airdropContractInfo = &AirdropContractInfo{
		airdropContract: airdropContract,
		auth:            clientInfo.auth,
		fromAddress:     clientInfo.fromAddress,
		client:          clientInfo.client,
	}

	return airdropContractInfo, nil
}

func GetAirdropContracts() *AirdropContractInfo {
	return airdropContractInfo
}

// GetClient 获取区块链客户端（只读）
func (c *AirdropContractInfo) GetClient() *ethclient.Client {
	return c.client
}

// GetFromAddress 获取发起地址（只读）
func (c *AirdropContractInfo) GetFromAddress() common.Address {
	return c.fromAddress
}

// GetFromAddress 获取发起地址（只读）
func (c *AirdropContractInfo) GetNewContract() *airdrop.Contracts {
	return c.airdropContract
}

// GetAuth 获取带并发安全的签名器（返回副本，避免外部直接修改原 Auth）
// 若需修改 Nonce/Gas 等字段，通过此方法获取副本后修改，再通过 UpdateAuth 同步
func (c *AirdropContractInfo) GetAuth() *bind.TransactOpts {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 返回 Auth 的副本，避免外部直接修改原字段（深拷贝关键字段）
	authCopy := *c.auth
	// 若 Auth.Context 非 nil，可按需拷贝（避免上下文被外部关闭）
	if c.auth.Context != nil {
		authCopy.Context = c.auth.Context
	}
	return &authCopy
}

// UpdateAuth 并发安全地更新签名器（如更新 Nonce）
func (c *AirdropContractInfo) UpdateAuth(newAuth *bind.TransactOpts) error {
	if newAuth == nil {
		return fmt.Errorf("newAuth 不能为空")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 深拷贝新 Auth 的关键字段，避免外部引用污染
	c.auth.Nonce = newAuth.Nonce
	c.auth.GasPrice = newAuth.GasPrice
	c.auth.GasLimit = newAuth.GasLimit
	c.auth.Context = newAuth.Context
	return nil
}
