package model

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Username  string `gorm:"type:varchar(64);uniqueIndex;not null"`
	Password  string `gorm:"type:varchar(128);not null"`
	Email     string `gorm:"type:varchar(128);not null;default:''"`
	Nickname  string `gorm:"type:varchar(64);not null;default:''"`
	Avatar    string `gorm:"type:varchar(512);not null;default:''"`
	Bio       string `gorm:"type:varchar(255);not null;default:''"`
	Gender    int8   `gorm:"not null;default:0"`
	Phone     string `gorm:"type:varchar(20);not null;default:''"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
