package entity

import (
	"time"

	"gorm.io/gorm"
)

type Price struct {
	gorm.Model

	UserID   uint      `gorm:"not null,index"`
	DateTime time.Time `gorm:"not null"`
	Store    string    `gorm:"not null;size:255"`
	Product  string    `gorm:"not null;size:255"`
	Price    uint      `gorm:"not null"`
	InStock  bool      `gorm:"not null"`
}
