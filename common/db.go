package common

import "time"

type User struct {
	ID        int64  `gorm:"primary_key"`
	Category  string `gorm:"type:varchar(255);not null"`
	Tweets    []Tweet
	UpdatedAt time.Time
	CreatedAt time.Time
}

type Tweet struct {
	UserId    int64 `gorm:"index"`
	TweetId   int64
	Text      string `gorm:"type:varchar(140);not null"`
	Images    []Image
	UpdatedAt time.Time
	CreatedAt time.Time
}

type Image struct {
	TweetId   int64
	URL       string `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time
	CreatedAt time.Time
}
