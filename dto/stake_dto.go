package dto

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type StakeResponse struct {
	Hash            string         `json:"hash"`
	ContractAddress string         `json:"contractAddress"`
	FromAddress     common.Address `json:"fromAddress"`
	Method          string         `json:"method"`
}
type StakeRequest struct {
	Amount int64 `json:"amount"`
	Period uint8 `json:"period"`
}

type WithDrawnRequest struct {
	Index big.Int `json:"index"`
}
