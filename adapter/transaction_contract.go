package adapter

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

var (
	transactionContractInfo *TransactionContractInfo
)

type TransactionContractInfo struct {
	FromAddress common.Address
	Client      *ethclient.Client
	Auth        *bind.TransactOpts
	ChainID     *big.Int
	PrivateKey  *ecdsa.PrivateKey
}

func NewTransactionContract() (*TransactionContractInfo, error) {
	clientInfo := GetInitClient()

	if clientInfo.Auth == nil {
		return nil, fmt.Errorf("auth should not be nil")
	}
	if clientInfo.Client == nil {
		return nil, fmt.Errorf("client should not be nil")
	}
	transactionContractInfo = &TransactionContractInfo{
		Auth:        clientInfo.Auth,
		FromAddress: clientInfo.FromAddress,
		ChainID:     clientInfo.ChainID,
		PrivateKey:  clientInfo.PrivateKey,
		Client:      clientInfo.Client,
	}
	return transactionContractInfo, nil
}
