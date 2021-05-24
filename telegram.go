package main

import (
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
	_, err := t.Bot.Send(t.Channel, message)
	return err
}
