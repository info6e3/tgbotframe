package tgbotframe

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Handler interface {
	Handle(bot *Bot, message *tgbotapi.Message) (tgbotapi.Chattable, error)
}

// HandlerWithMiddlewares

type HandlerWithMiddlewares struct {
	Handler     Handler
	Middlewares []Middleware
}

func (h *HandlerWithMiddlewares) applyMiddlewares(bot *Bot, message *tgbotapi.Message) error {
	for _, v := range h.Middlewares {
		err := v.Apply(bot.api, message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *HandlerWithMiddlewares) Handle(bot *Bot, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	if err := h.applyMiddlewares(bot, message); err != nil {
		return nil, err
	}

	ch, err := h.Handler.Handle(bot, message)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// HandleFunc

type HandleFunc struct {
	Func func(bot *Bot, message *tgbotapi.Message) (tgbotapi.Chattable, error)
}

func (h *HandleFunc) Handle(bot *Bot, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	return h.Func(bot, message)
}
