package main

import (
	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type TwitterConfig struct {
	ConsumerKey, ConsumerSecret, AccessToken, AccessSecret string
}

type TwitterAPI struct {
	Client *twitter.Client
}

func (t *TwitterAPI) Start() error {
	consumerKey, err := MustGetString("twitter.consumerKey")
	if err != nil {
		return err
	}
	consumerSecret, err := MustGetString("twitter.consumerSecret")
	if err != nil {
		return err
	}
	accessToken, err := MustGetString("twitter.accessToken")
	if err != nil {
		return err
	}
	accessSecret, err := MustGetString("twitter.accessSecret")
	if err != nil {
		return err
	}
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	t.Client = twitter.NewClient(config.Client(oauth1.NoContext, token))
	return nil
}

func (t *TwitterAPI) Post(message string) error {
	log.Printf("Total Twitter message length: %d", len(message))
	if len(message) > 280 {
		log.Printf("Message too long for Twitter :(")
	}
	t.Client.Statuses.Update(message, &twitter.StatusUpdateParams{})
	return nil
}
