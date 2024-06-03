package models

import (
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/DanillaY/GoScrapper/cmd/repository"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type Book struct {
	ID               uint `gorm:"primaryKey;autoIncrement:true"`
	CurrentPrice     int
	OldPrice         int
	Title            string
	ImgPath          string
	PageBookPath     string // absolute url of the book page example -> https://book24.ru/product/edinstvennyy-i-ego-sobstvennost-5386506/
	VendorURL        string // url of the site example -> https://book24.ru/
	Vendor           string
	Author           string
	Translator       string
	ProductionSeries string
	Category         string
	Publisher        string
	ISBN             string
	AgeRestriction   string
	YearPublish      string
	PagesQuantity    string
	BookCover        string
	Format           string
	Weight           string
	InStockText      string
	BookAbout        string
	User             []*User `gorm:"many2many:book_users;"`
}

func (b *Book) BeforeUpdate(tx *gorm.DB) error {

	config, err := repository.GetConfigVariables()

	if tx.Statement.Changed("is_in_stock") && err == nil && b.InStockText != "В наличии" {
		for _, user := range b.User {

			fmt.Println("\n Зашел в бефор " +
				strconv.FormatBool(tx.Statement.Changed("is_in_stock")) +
				strconv.Itoa(len(b.User)) + "\n" +
				" test: " + b.InStockText)

			m := gomail.NewMessage()

			m.SetHeader("From", config.EMAIL_USERNAME)
			m.SetHeader("To", user.Email)
			m.SetHeader("Subject", "NEW BOOK JUST DROPPED!!")
			m.SetBody("text/plain", "A book that you added to the favorite just appeared! \n"+b.Title+" "+b.PageBookPath)

			d := gomail.NewDialer("smtp.yandex.ru", 587, config.EMAIL_USERNAME, config.EMAIL_PASSWORD)
			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

			if err = d.DialAndSend(m); err != nil {
				fmt.Println("Could not send email")
				return err
			}
		}
	}
	return nil
}
