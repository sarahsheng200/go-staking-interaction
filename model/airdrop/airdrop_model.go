package airdrop

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Wallet struct {
	PrivateKey string // 私钥（十六进制）
	Address    string // 以太坊地址（0x前缀）
}

type Request struct {
	Count     int      `json:"count"`
	BatchSize int      `json:"batchSize"`
	Amount    *big.Int `json:"amount"`
}

type Response struct {
	BatchNum        int              `json:"batchNum"`
	Hash            string           `json:"hash"`
	ContractAddress string           `json:"contractAddress"`
	FromAddress     string           `json:"fromAddress"`
	WalletAddress   []common.Address `json:"walletAddress"`
	Error           string           `json:"error"`
}
