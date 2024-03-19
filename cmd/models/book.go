package models

type Book struct {
	ID               uint `gorm:"primaryKey;autoIncrement:true"`
	CurrentPrice     int
	OldPrice         int
	Title            string
	ImgPath          string
	PageBookPath     string // absolute url of the book page example -> https://book24.ru/product/edinstvennyy-i-ego-sobstvennost-5386506/
	Vendor           string // url of the site example -> https://book24.ru/
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
	BookAbout        string
}
