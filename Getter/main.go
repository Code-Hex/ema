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
	"github.com/k0kubun/pp"
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
	watson.APIStreaming()
}

func (watson *Watson) APIStreaming() {

	client := watson.Tw.Auth()

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = watson.StreamAndInsert

	fmt.Println("Starting Stream...")

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"猫", "ねこ", "にゃんこ", "にゃー"},
		StallWarnings: twitter.Bool(true),
		Language:      []string{"ja"},
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	// USER (quick test: auth'd user likes a tweet -> event)
	// userParams := &twitter.StreamUserParams{
	// 	StallWarnings: twitter.Bool(true),
	// 	With:          "followings",
	// 	Language:      []string{"en"},
	// }
	// stream, err := client.Streams.User(userParams)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// SAMPLE
	// sampleParams := &twitter.StreamSampleParams{
	// 	StallWarnings: twitter.Bool(true),
	// }
	// stream, err := client.Streams.Sample(sampleParams)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	fmt.Println("Stopping Stream...")
	stream.Stop()
}

func (watson *Watson) StreamAndInsert(tweet *twitter.Tweet) {
	if tweet.Lang == "ja" && tweet.RetweetedStatus == nil {
		pp.Println(tweet.Text)
		tweetdb := new(common.Tweet)
		userdb := new(common.User)
		userdb.ID = tweet.User.ID
		tweetdb.Text = tweet.Text

		imagedb := make([]common.Image, 0, 4)
		for _, media := range tweet.Entities.Media {
			if media.Type == "photo" {
				imagedb = append(imagedb, common.Image{
					TweetId:   tweet.ID,
					URL:       media.MediaURLHttps,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				})
			}
		}

		tweetdb.Images = imagedb
		fmt.Printf("Inserted: %d\n", userdb.ID)
		watson.DB.Create(&userdb)
		watson.DB.Create(&tweetdb)
	}
}
