package main

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type User struct {
	ID        uint   `gorm:"primary_key"`
	Category  string `gorm:"type:varchar(255);not null"`
	Tweets    []Tweet
	CreatedAt time.Time
}

type Tweet struct {
	UserId    uint   `gorm:"index"`
	TweetId   string `gorm:"type:varchar(255);not null"`
	Text      string `gorm:"type:varchar(140);not null"`
	ImageURL  string `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time
}

func main() {
	db, err := gorm.Open("postgres", "host=localhost dbname=test sslmode=disable")
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&User{}, &Tweet{})
	defer db.Close()
}
