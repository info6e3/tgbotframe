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
	token          string
	noneStop       bool
	api            *tgbotapi.BotAPI
	logMutex       *sync.Mutex
	middlewares    []Middleware
	textHandler    Handler
	cmdHandlers    map[string]Handler
	voiceHandler   Handler
	customHandlers Handler
	recipients     []int64
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

	if err := b.applyMiddlewares(message); err != nil {
		return
	}

	switch {
	case message.Text != "":
		// Обработка команд /
		if strings.HasPrefix(message.Text, "/") {
			if len(b.cmdHandlers) > 0 {
				text := strings.TrimPrefix(message.Text, "/")
				prefixAndText := strings.SplitN(text, " ", 2)
				if len(prefixAndText) == 2 {
					prefix := prefixAndText[0]
					if b.cmdHandlers[prefix] != nil {
						msg, err := b.cmdHandlers[prefix].Handle(b, message)
						if err == nil && msg != nil {
							b.send(msg)
						}
					}
				}
			}
		} else if b.textHandler != nil { // Обработка текста
			if !message.Chat.IsGroup() && !message.Chat.IsSuperGroup() {
				msg, err := b.textHandler.Handle(b, message)
				if err == nil && msg != nil {
					b.send(msg)
				}
			}
		}
		// Обработка войсов
	case message.Voice != nil:
		msg, err := b.voiceHandler.Handle(b, message)
		if err == nil && msg != nil {
			b.send(msg)
		}
	}

	// TODO: Вынести отдельно все дополнительные

	if len(b.recipients) > 0 {
		for _, recipient := range b.recipients {
			msg := tgbotapi.NewCopyMessage(recipient, message.Chat.ID, message.MessageID)
			b.send(msg)
		}
	}
}

// Bot Functions

func (b *Bot) send(chattable tgbotapi.Chattable) {
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

func (b *Bot) applyMiddlewares(message *tgbotapi.Message) error {
	for _, v := range b.middlewares {
		err := v.Apply(b.api, message)
		if err != nil {
			return err
		}
	}
	return nil
}

// Bot Handlers

func (b *Bot) SetTextHandler(handler Handler) {
	b.textHandler = handler
}

func (b *Bot) SetVoiceHandler(handler Handler) {
	b.voiceHandler = handler
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
