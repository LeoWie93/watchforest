package models

import "gorm.io/gorm"

type Session struct {
	gorm.Model
	UserID      uint
	SessionID   int
	AccessToken string
	ExpireDate  int
}
