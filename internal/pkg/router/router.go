package router

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type receiver interface {
	Receive(message *tgbotapi.Message)
}

type impl struct {
	bot  receiver
	chat receiver
}

// New конструктор роутера.
func New(bot receiver, chat receiver) *impl {
	return &impl{
		bot:  bot,
		chat: chat,
	}
}

// Route роутит сообщения в зависимости от вызываемого контекста.
func (i *impl) Route(updates <-chan tgbotapi.Update) {
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// если вызвали из чата
		r := i.chat

		// если вызвали в личку бота
		if update.Message.Chat.IsPrivate() {
			r = i.bot
		}

		r.Receive(update.Message)
	}
}
