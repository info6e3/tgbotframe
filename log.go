package tgbotframe

import (
	"fmt"
	"log"
	"os"
)

func (b *Bot) log(data []byte) {
	b.logMutex.Lock()
	f, err := os.OpenFile("log.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer func() {
		f.Close()
		b.logMutex.Unlock()
	}()

	_, err = f.Write(data)
	if err != nil {
		log.Println(err)
	}

	_, err = f.Write([]byte(fmt.Sprintln()))
	if err != nil {
		log.Println(err)
	}
}
