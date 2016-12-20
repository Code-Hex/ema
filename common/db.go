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
	ID        int64  `gorm:"primary_key"`
	UserID    int64  `gorm:"primary_key"`
	Text      string `gorm:"type:varchar(255);not null"`
	Images    []Image
	UpdatedAt time.Time
	CreatedAt time.Time
}

type Image struct {
	ID        int64
	URL       string `gorm:"type:varchar(255);not null"`
	UpdatedAt time.Time
	CreatedAt time.Time
}
