package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

type impl struct {
	bot   *tgbotapi.BotAPI
	errCh chan<- error
}

func New(bot *tgbotapi.BotAPI, errCh chan<- error) *impl {
	return &impl{
		bot:   bot,
		errCh: errCh,
	}
}

func (i *impl) Receive(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)
	msg.ReplyToMessageID = message.MessageID

	_, err := i.bot.Send(msg)
	if err != nil {
		i.errCh <- errors.Wrap(err, "cannot send message")
	}
}
