package models

import (
	"time"
)

type Dao struct {
	ID         string     `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	ChainID    int        `gorm:"column:chain_id;not null" json:"chain_id"`
	ChainName  string     `gorm:"column:chain_name;type:varchar(255);not null" json:"chain_name"`
	Name       string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Code       string     `gorm:"column:code;type:varchar(255);not null;uniqueIndex:uq_dgv_dao_code" json:"code"`
	Seq        int        `gorm:"column:seq;not null;default:0" json:"seq"`
	State      string     `gorm:"column:state;type:varchar(50);not null" json:"state"`
	ConfigLink string     `gorm:"column:config_link;type:varchar(255);not null" json:"config_link"`
	TimeSyncd  *time.Time `gorm:"column:time_syncd" json:"time_syncd,omitempty"`
	CTime      time.Time  `gorm:"column:ctime;default:now()" json:"ctime"`
	UTime      *time.Time `gorm:"column:utime" json:"utime,omitempty"`
}

func (Dao) TableName() string {
	return "dgv_dao"
}
