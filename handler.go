package tgbotframe

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Handler interface {
	Handle(bot *Bot, message *tgbotapi.Message) (ok bool)
}

// HandlerWithMiddlewares

type HandlerWithMiddlewares struct {
	Handler     Handler
	Middlewares []Middleware
}

func (h *HandlerWithMiddlewares) applyMiddlewares(bot *Bot, message *tgbotapi.Message) (ok bool) {
	for _, v := range h.Middlewares {
		if ok = v.Apply(bot.api, message); ok == false {
			return false
		}
	}
	return true
}

func (h *HandlerWithMiddlewares) Handle(bot *Bot, message *tgbotapi.Message) (ok bool) {
	if ok = h.applyMiddlewares(bot, message); ok == false {
		return false
	}

	if ok = h.Handler.Handle(bot, message); ok == false {
		return false
	}

	return true
}

// HandleFunc

type HandleFunc struct {
	Func func(bot *Bot, message *tgbotapi.Message) (ok bool)
}

func (h *HandleFunc) Handle(bot *Bot, message *tgbotapi.Message) (ok bool) {
	return h.Func(bot, message)
}
