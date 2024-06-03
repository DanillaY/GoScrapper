package models

type User struct {
	ID    uint `gorm:"primaryKey;autoIncrement:true"`
	Email string
	Book  []*Book `gorm:"many2many:book_users;"`
}
