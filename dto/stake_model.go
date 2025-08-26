package dto

import (
	"math/big"
)

type StakeResponse struct {
	Hash            string `json:"hash"`
	ContractAddress string `json:"contractAddress"`
	FromAddress     string `json:"fromAddress"`
	Method          string `json:"method"`
}
type StakeRequest struct {
	Amount int64 `json:"amount"`
	Period uint8 `json:"period"`
}

type WithDrawnRequest struct {
	Index big.Int `json:"index"`
}
