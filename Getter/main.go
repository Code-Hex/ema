package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Code-Hex/ema/common"
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
	watson.Crawl()
}

func (watson *Watson) Crawl() {

	client := watson.Tw.Auth()

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = watson.FetchUserInsert

	fmt.Println("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"猫", "ねこ", "ネコ", "にゃんこ", "にゃー"},
		StallWarnings: twitter.Bool(true),
		Language:      []string{"ja"},
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	go watson.CrawlTimeline()

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	log.Println(<-ch)

	fmt.Println("Stopping Stream...")
	stream.Stop()
}

func (watson *Watson) CrawlTimeline() {
	rows, err := watson.DB.Model(&common.User{}).Select("id").Rows()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var user common.User
		watson.DB.ScanRows(rows, &user)
		watson.GetUserTimeline(user.ID)
	}
}

func (watson *Watson) GetUserTimeline(id int64) {
	var tweetid int64
	client := watson.Tw.Auth()
	for {
		count := 100

		for i := 1; i < 10; i++ {
			tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
				UserID: id,
				MaxID:  tweetid,
				Count:  count,
			})

			if tweets == nil {
				return
			}

			if err != nil {
				log.Println("Error:", err.Error())
				continue
			}
			watson.InsertUserTweets(tweets)
			tweetid = tweets[len(tweets)-1].ID
			log.Println("MaxID:", tweetid)
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

func (watson *Watson) FetchUserInsert(tweet *twitter.Tweet) {
	if tweet.Lang == "ja" && tweet.RetweetedStatus == nil {
		userdb := new(common.User)

		var count int64
		watson.DB.Where("id = ?", tweet.User.ID).First(&common.User{}).Count(&count)
		if count == 0 {
			userdb.ID = tweet.User.ID
			fmt.Printf("Inserted: %d\n", userdb.ID)
			watson.DB.Create(userdb)
		}
	}
}
