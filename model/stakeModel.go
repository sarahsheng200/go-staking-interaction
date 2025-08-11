package model

import (
	"math/big"
	"time"
)

type Response struct {
	Hash            string  `json:"hash"`
	ContractAddress string  `json:"contractAddress"`
	FromAddress     string  `json:"fromAddress"`
	GasUsed         float64 `json:"gasUsed"`
	Method          string  `json:"method"`
}
type StakeRequest struct {
	Amount int64 `json:"amount"`
	Period uint8 `json:"period"`
}

type WithDrawnRequest struct {
	Index big.Int `json:"index"`
}

// Stake 质押记录表对应的结构体
type Stake struct {
	ID              int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement" comment:"自增主键ID"`
	IndexNum        big.Int   `json:"index_num" gorm:"column:index_num;not null" comment:"索引编号"`
	Hash            string    `json:"hash" gorm:"column:hash;type:varchar(100);not null;uniqueIndex:uk_hash" comment:"交易哈希"`
	ContractAddress string    `json:"contract_address" gorm:"column:contract_address;type:varchar(100);not null" comment:"合约地址"`
	FromAddress     string    `json:"from_address" gorm:"column:from_address;type:varchar(100);not null" comment:"发起地址"`
	Method          string    `json:"method" gorm:"column:method;type:varchar(20);not null" comment:"操作方法：stake-质押，withdraw-提取"`
	Amount          float64   `json:"amount" gorm:"column:amount;type:decimal(20,8)" comment:"交易金额"`
	GasUsed         float64   `json:"gas_used" gorm:"column:gas_used;type:decimal(10,2);default:0.00" comment:"消耗的Gas"`
	BlockNumber     int64     `json:"block_number" gorm:"column:block_number;type:bigint" comment:"区块编号"`
	Status          int8      `json:"status" gorm:"column:status;type:tinyint;default:0" comment:"状态：0-质押中，1-已提取"`
	Timestamp       time.Time `json:"timestamp" gorm:"column:timestamp;type:datetime" comment:"交易时间戳"`
	CreatedDate     time.Time `json:"created_date" gorm:"column:created_date;default:current_timestamp" comment:"记录创建时间"`
	UpdatedDate     time.Time `json:"updated_date" gorm:"column:updated_date;default:current_timestamp on update current_timestamp" comment:"记录更新时间"`
}
