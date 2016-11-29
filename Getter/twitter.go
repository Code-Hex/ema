package main

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Twitter struct {
	ConsumerKey    string `envconfig:"TWITTER_CONSUMER_KEY"`
	ConsumerSecret string `envconfig:"TWITTER_CONSUMER_SECRET"`
	AccessToken    string `envconfig:"TWITTER_ACCESS_TOKEN"`
	AccessSecret   string `envconfig:"TWITTER_ACCESS_SECRET"`
}

func (twitter Twitter) HasData() bool {
	consumerKey := twitter.ConsumerKey
	consumerSecret := twitter.ConsumerSecret
	accessToken := twitter.AccessToken
	accessSecret := twitter.AccessSecret
	return consumerKey != "" && consumerSecret != "" && accessToken != "" && accessSecret != ""
}

func (tw Twitter) Auth() *twitter.Client {
	config := oauth1.NewConfig(tw.ConsumerKey, tw.ConsumerSecret)
	token := oauth1.NewToken(tw.AccessToken, tw.AccessSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}
