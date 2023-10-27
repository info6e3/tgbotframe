package tgbotframe

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Middleware interface {
	Apply(*tgbotapi.BotAPI, *tgbotapi.Message) (ok bool)
}
