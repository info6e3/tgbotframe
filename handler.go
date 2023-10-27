package tgbotframe

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Handler interface {
	Handle(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (tgbotapi.Chattable, error)
}

// HandlerWithMiddlewares

type HandlerWithMiddlewares struct {
	Handler     Handler
	Middlewares []Middleware
}

func (h *HandlerWithMiddlewares) applyMiddlewares(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	for _, v := range h.Middlewares {
		err := v.Apply(bot, message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *HandlerWithMiddlewares) Handle(bot *tgbotapi.BotAPI, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	if err := h.applyMiddlewares(bot, message); err != nil {
		return nil, err
	}

	ch, err := h.Handler.Handle(bot, message)
	if err != nil {
		return nil, err
	}

	return ch, nil
}
