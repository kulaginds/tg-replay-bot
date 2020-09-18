package chat

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

type searchEngine interface {
	AddQuery(query string)
	HasQueries(text string) bool
}

type impl struct {
	bot   *tgbotapi.BotAPI
	errCh chan<- error

	chatID int64

	searchEngine searchEngine
}

func New(bot *tgbotapi.BotAPI, errCh chan<- error, chatID int64, searchEngine searchEngine) *impl {
	searchEngine.AddQuery("hello")
	searchEngine.AddQuery("hey world")

	return &impl{
		bot:          bot,
		errCh:        errCh,
		chatID:       chatID,
		searchEngine: searchEngine,
	}
}

func (i *impl) Receive(message *tgbotapi.Message) {
	if !i.searchEngine.HasQueries(message.Text) {
		return
	}

	msg := tgbotapi.NewForward(i.chatID, message.Chat.ID, message.MessageID)

	_, err := i.bot.Send(msg)
	if err != nil {
		i.errCh <- errors.Wrap(err, "cannot send message")
	}
}
