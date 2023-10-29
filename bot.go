package tgbotframe

import (
	"encoding/json"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"sync"
)

// Structure

type Bot struct {
	token       string
	noneStop    bool
	api         *tgbotapi.BotAPI
	logMutex    *sync.Mutex
	middlewares []Middleware
	cmdHandlers map[string]Handler
	handlers    []Handler
	recipients  []int64
}

func NewBot(token string, noneStop bool) *Bot {
	return &Bot{
		token:    token,
		noneStop: noneStop,
	}
}

/* TODO: Remove logic in the constructor. Maybe add in Middleware interface func Init()? */

func (b *Bot) Run() {
	var err error
	b.api, err = tgbotapi.NewBotAPI(b.token)
	if err != nil {
		log.Println(err)
	}

	b.logMutex = &sync.Mutex{}

	if b.noneStop {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
				b.api.StopReceivingUpdates()
				b.Run()
			}
		}()
	}

	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go b.handle(update.Message)
		}
	}
}

func (b *Bot) handle(message *tgbotapi.Message) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Panic occurred:", err)
		}
	}()

	jsonData, _ := json.Marshal(message)
	b.log(jsonData)

	if ok := b.applyMiddlewares(message); ok == false {
		return
	}

	// Обработка команд /
	if message.Text != "" {
		if strings.HasPrefix(message.Text, "/") {
			if len(b.cmdHandlers) > 0 {
				text := strings.TrimPrefix(message.Text, "/")
				prefixAndText := strings.SplitN(text, " ", 2)
				prefix := prefixAndText[0]
				msgWithoutPrefix := message
				if len(prefixAndText) == 2 {
					msgWithoutPrefix.Text = prefixAndText[1]
				} else {
					msgWithoutPrefix.Text = ""
				}
				b.cmdHandlers[prefix].Handle(b, msgWithoutPrefix)
			}
		}
	}

	for _, handler := range b.handlers {
		handler.Handle(b, message)
	}

	// TODO: Вынести отдельно все дополнительные

	for _, recipient := range b.recipients {
		msg := tgbotapi.NewCopyMessage(recipient, message.Chat.ID, message.MessageID)
		b.Send(msg)
	}
}

// Bot Functions

func (b *Bot) Send(chattable tgbotapi.Chattable) {
	jsonData, _ := json.Marshal(chattable)
	b.log(jsonData)
	_, err := b.api.Send(chattable)
	if err != nil {
		log.Println(err)
	}
}

// Bot Middlewares

func (b *Bot) SetMiddlewares(middlewares []Middleware) {
	b.middlewares = middlewares
}

func (b *Bot) applyMiddlewares(message *tgbotapi.Message) (ok bool) {
	for _, v := range b.middlewares {
		if ok = v.Apply(b.api, message); ok == false {
			return ok
		}
	}
	return true
}

// Bot Handlers

func (b *Bot) SetHandler(handler Handler) {
	b.handlers = append(b.handlers, handler)
}

func (b *Bot) SetCmdHandler(key string, handler Handler) {
	/* TODO: Проверить мап */
	if b.cmdHandlers == nil {
		b.cmdHandlers = make(map[string]Handler)
	}
	b.cmdHandlers[key] = handler
}

// Bot additional functions

func (b *Bot) SetRecipient(chatId int64) {
	b.recipients = append(b.recipients, chatId)
}

func (b *Bot) RemoveRecipient(chatId int64) {
	for i, recipient := range b.recipients {
		if recipient == chatId {
			size := len(b.recipients)
			b.recipients[i] = b.recipients[size-1]
			b.recipients = b.recipients[:size-1]
		}
	}
}
