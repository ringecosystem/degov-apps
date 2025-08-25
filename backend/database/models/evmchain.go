package dbmodels

import "time"

type ContractsAbiType string

const (
	ContractsAbiTypeProxy          ContractsAbiType = "PROXY"
	ContractsAbiTypeImplementation ContractsAbiType = "IMPLEMENTATION"
)

type ContractsAbi struct {
	ID             string           `gorm:"column:id;type:varchar(128);primaryKey" json:"id"`
	ChainId        int              `gorm:"column:chain_id;not null" json:"chain_id"`
	Address        string           `gorm:"column:address;type:varchar(255);not null" json:"address"`
	Type           ContractsAbiType `gorm:"column:type;type:varchar(30);not null" json:"type"`
	Implementation string           `gorm:"column:implementation;type:varchar(255);" json:"implementation"`
	Abi            string           `gorm:"column:abi;type:text;" json:"abi"`
	CTime          time.Time        `gorm:"column:ctime;default:now()" json:"ctime"`
	UTime          *time.Time       `gorm:"column:utime" json:"utime,omitempty"`
}

func (ContractsAbi) TableName() string {
	return "dgv_contracts_abi"
}
