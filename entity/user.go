package entity

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Name     string `gorm:"not null;uniqueIndex;size:255"`
	Password string `gorm:"not null;size:255"`
}
