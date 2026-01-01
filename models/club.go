package models

import "gorm.io/gorm"

type Club struct {
	gorm.Model
	Name    string  `gorm:"size:100;not null"`
	City    string  `gorm:"size:100"`
	Country string  `gorm:"size:100"`
}

func (Club) TableName() string {
	return "clubs"
}
