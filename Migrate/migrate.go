package main

import (
	"github.com/Code-Hex/ema/common"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	db, err := gorm.Open("postgres", "host=localhost dbname=test sslmode=disable")
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&common.User{}, &common.Tweet{}, &common.Image{})
	defer db.Close()
}
