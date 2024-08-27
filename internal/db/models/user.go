package models

type User struct {
	ID           int    `gorm:"primary_key"`
	CernPersonId string `gorm:"type:varchar(255);not null;unique"`
	Username     string `gorm:"type:varchar(255);not null"`
	FirstName    string `gorm:"type:varchar(255);not null"`
	FamilyName   string `gorm:"type:varchar(255);not null"`
	Email        string `gorm:"type:varchar(255);unique;not null"`
}
