package common

import "staking-interaction/token"

const (
	PRIVATE_KEY              = token.LOCAL_PRIVATE_KEY
	STAKE_CONTRACT_ADDRESS   = token.LOCAL_STAKE_CONTRACT_ADDRESS
	AIRDROP_CONTRACT_ADDRESS = token.AIRDROP_CONTRACT_ADDRESS
	TOKEN_CONTRACT_ADDRESS   = token.TOKEN_CONTRACT_ADDRESS
	MYSQL_USERNAME           = "root"
	MYSQL_PASSWORD           = "00000000"
	MYSQL_DATABASE           = "web3-contract"
	MYSQL_URL                = "127.0.0.1:3306"
	MYSQL_CONFIG             = "charset=utf8&parseTime=True&loc=Local"
	RAW_URL                  = "https://data-seed-prebsc-2-s1.binance.org:8545"
	RPC_URL                  = "https://bsc-testnet-rpc.publicnode.com"
)

const (
	StakedEventName    = "Staked"
	WithdrawnEventName = "Withdrawn"
	BatchSize          = 100
)

const (
	TokenTypeBNB       = 1
	TokenTypeMTK       = 2
	BillTypeRecharge   = 1
	BillTypeWithdrawal = 2
	SyncBlockBuffer    = 10
)

var OwnerAddresses = []string{
	token.OWNER1,
	token.OWNER2,
	token.OWNER3,
	token.OWNER4,
	token.OWNER5,
}
