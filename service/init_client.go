package service

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

var (
	clientInfo InitClient
)

type InitClient struct {
	Auth        *bind.TransactOpts
	FromAddress string
	Client      *ethclient.Client
	PrivateKey  *ecdsa.PrivateKey
	ChainID     *big.Int
}

func GetInitClient() InitClient {
	return clientInfo
}

func NewInitClient(c InitClient) {
	clientInfo = InitClient{
		Auth:        c.Auth,
		FromAddress: c.FromAddress,
		Client:      c.Client,
		PrivateKey:  c.PrivateKey,
		ChainID:     c.ChainID,
	}
}
