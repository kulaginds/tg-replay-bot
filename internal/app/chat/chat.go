package chat

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

type impl struct {
	bot   *tgbotapi.BotAPI
	errCh chan<- error

	chatID int64
}

func New(bot *tgbotapi.BotAPI, errCh chan<- error, chatID int64) *impl {
	return &impl{
		bot:    bot,
		errCh:  errCh,
		chatID: chatID,
	}
}

func (i *impl) Receive(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(i.chatID, message.Text)
	//msg := tgbotapi.NewMessageToChannel("@tg_reply_bot_privcha", update.Message.Text)
	//msg.ReplyToMessageID = update.Message.MessageID

	_, err := i.bot.Send(msg)
	if err != nil {
		i.errCh <- errors.Wrap(err, "cannot send message")
	}
}
