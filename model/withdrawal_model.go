package model

type Withdrawal struct {
	ID            uint64 `gorm:"column:id;type:bigint unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	TokenType     int    `gorm:"column:token_type;type:int" json:"token_type"` // 1.BNB 2.ROM 3.WCN
	WalletAddress string `gorm:"column:wallet_address;type:varchar(64)" json:"wallet_address"`
	Amount        string `gorm:"column:amount;type:varchar(64)" json:"amount"` // 数量, 10bnb
	Value         string `gorm:"column:value;type:varchar(64)" json:"value"`   // 实际到账数量, 9bnb
	Fee           string `gorm:"column:fee;type:varchar(64)" json:"fee"`       // 手续费, 1bnb
	GasPrice      string `gorm:"column:gas_price;type:varchar(64);default:'0'" json:"gas_price"`
	Status        int8   `gorm:"column:status;type:tinyint" json:"status"` // 1.INIT 2.PASS 3.REJECT 4.PENDING 5.SUCCESS 6.FAIL 7.ABNORMAL
	Hash          string `gorm:"column:hash;type:varchar(120)" json:"hash"`
	Nonce         int    `gorm:"column:nonce;type:int" json:"nonce"`
}
