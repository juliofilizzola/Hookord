package telegram

import (
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramClient interface {
	SendMessage(chatId string, text string) (int, error)
	EditMessage(chatId string, messageId int, text string) (int, error)
	Close() error
}

type sessionClient struct {
	bot *tgbotapi.BotAPI
}

func NewSessionClient(token string) (TelegramClient, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, ErrSessionConnect
	}
	return &sessionClient{bot: bot}, nil
}

func (c *sessionClient) SendMessage(chatId string, text string) (int, error) {
	id, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		return 0, err
	}
	msg := tgbotapi.NewMessage(id, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	sentMsg, err := c.bot.Send(msg)
	if err != nil {
		return 0, err
	}
	return sentMsg.MessageID, nil
}

func (c *sessionClient) EditMessage(chatId string, messageId int, text string) (int, error) {
	id, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		return 0, err
	}
	msg := tgbotapi.NewEditMessageText(id, messageId, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.DisableWebPagePreview = true
	sentMsg, err := c.bot.Send(msg)
	if err != nil {
		return 0, err
	}
	return sentMsg.MessageID, nil
}

func (c *sessionClient) Close() error {
	return nil
}
