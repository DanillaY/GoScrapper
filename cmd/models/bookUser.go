package models

type BookUser struct {
	UserID int `gorm:"primaryKey"`
	BookID int `gorm:"primaryKey"`
}
