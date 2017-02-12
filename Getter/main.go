package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Code-Hex/ema/common"
	"github.com/Songmu/prompter"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kelseyhightower/envconfig"
)

type Watson struct {
	Tw Twitter
	DB *gorm.DB
}

func New() *Watson {
	w := new(Watson)
	if err := envconfig.Process("twitter", &w.Tw); err != nil {
		log.Fatal(err.Error())
	}

	db, err := gorm.Open("postgres", "host=localhost dbname=test sslmode=disable")
	if err != nil {
		panic(err)
	}
	w.DB = db

	return w
}

func (w *Watson) Close() error {
	return w.DB.Close()
}

func (watson *Watson) HasData() bool {
	return watson.Tw.HasData()
}

func main() {
	watson := New()
	defer watson.Close()

	if !watson.HasData() {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}
	watson.UserInput()
	log.Println("Start Crawl...")
	watson.Crawl()
}

func (watson *Watson) Crawl() {
	go watson.CrawlTimeline()

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	log.Println(<-ch)
}

func (watson *Watson) UserInput() {
	client := watson.Tw.Auth()
	for {
		screenID := prompter.Prompt("Enter user twitter ID", "")
		if screenID == "" {
			break
		}

		params := &twitter.UserShowParams{
			ScreenName:      screenID,
			IncludeEntities: twitter.Bool(false),
		}
		user, _, err := client.Users.Show(params)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		userdb := new(common.User)

		var count int64
		watson.DB.Where("id = ?", user.ID).First(new(common.User)).Count(&count)
		if count == 0 {
			userdb.ID = user.ID
			fmt.Printf("Inserted: %d\n", userdb.ID)
			watson.DB.Create(userdb)
		} else {
			log.Printf("Already exist: %s as %d\n", user.ScreenName, user.ID)
		}
	}
}

func (watson *Watson) CrawlTimeline() {
	rows, err := watson.DB.Model(new(common.User)).Select("id").Rows()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var (
			user  common.User
			tweet common.Tweet
		)
		watson.DB.ScanRows(rows, &user)
		watson.DB.Where("user_id = ?", user.ID).First(&tweet)
		watson.GetUserTimeline(user.ID, tweet.ID)
	}
}

func (watson *Watson) GetUserTimeline(id, tid int64) {
	client := watson.Tw.Auth()

	for {
		count := 100

		for i := 1; i < 10; i++ {
			tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
				UserID: id,
				MaxID:  tid,
				Count:  count,
			})
			if err != nil {
				log.Println("Error:", err.Error())
				continue
			}

			if len(tweets) <= 1 {
				return
			}

			watson.InsertUserTweets(tweets)
			tid = tweets[len(tweets)-1].ID
			log.Println("MaxID:", tid)
		}

		log.Println("Sleeping...")
		time.Sleep(16 * time.Minute)

	}
}

func (watson *Watson) InsertUserTweets(tweets []twitter.Tweet) {
	var count int64
	log.Printf("Insert data count: %d\n", len(tweets))

	for _, t := range tweets {
		tweetdb := new(common.Tweet)

		watson.DB.Where("id = ?", t.ID).First(&common.Tweet{}).Count(&count)
		if count > 0 {
			continue
		}

		tweetdb.ID = t.ID
		tweetdb.UserID = t.User.ID
		tweetdb.Text = t.Text
		for _, media := range t.Entities.Media {
			if media.Type == "photo" {
				watson.DB.Create(&common.Image{
					URL:       media.MediaURLHttps,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				})
			}
		}
		watson.DB.Create(tweetdb)
	}
}
