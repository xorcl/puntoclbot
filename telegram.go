package main

import (
	"log"

	"gopkg.in/tucnak/telebot.v2"
)

type TelegramBot struct {
	Bot     *telebot.Bot
	Channel *telebot.Chat
}

func (t *TelegramBot) Start() error {
	token, err := MustGetString("telegram.token")
	if err != nil {
		return err
	}
	chatID, err := MustGetString("telegram.chatID")
	if err != nil {
		return err
	}
	t.Bot, err = telebot.NewBot(telebot.Settings{
		Token: token,
	})
	if err != nil {
		return err
	}
	t.Channel, err = t.Bot.ChatByID(chatID)
	if err != nil {
		return err
	}
	return nil
}

func (t *TelegramBot) Post(message string) error {
	log.Printf("Total Twitter message length: %d", len(message))
	if len(message) > 280 {
		log.Printf("Message too long for Twitter :(")
	}
	_, err := t.Bot.Send(t.Channel, message, telebot.Silent)
	return err
}
