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

	if clientInfo.auth == nil {
		return nil, fmt.Errorf("auth should not be nil")
	}
	if clientInfo.client == nil {
		return nil, fmt.Errorf("client should not be nil")
	}
	transactionContractInfo = &TransactionContractInfo{
		Auth:        clientInfo.auth,
		FromAddress: clientInfo.fromAddress,
		ChainID:     clientInfo.chainID,
		PrivateKey:  clientInfo.privateKey,
		Client:      clientInfo.client,
	}
	return transactionContractInfo, nil
}
