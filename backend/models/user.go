package models

type User struct {
	ID      string  `gorm:"column:id;type:varchar(50);primaryKey" json:"id"`
	Address string  `gorm:"column:address;type:varchar(255);not null;uniqueIndex:uq_dgv_user_address" json:"address"`
	Email   *string `gorm:"column:email;type:varchar(255)" json:"email,omitempty"`
}

func (User) TableName() string {
	return "dgv_user"
}
