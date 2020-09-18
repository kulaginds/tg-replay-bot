package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envTelegramSecretToken = "ENV_TELEGRAM_SECRET_TOKEN"
	envTelegramOffset      = "ENV_TELEGRAM_OFFSET"

	updateTimeout = 60 // seconds
)

func main() {
	logCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Encoding:         "console",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			TimeKey:    "ts",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
		},
		DisableStacktrace: true,
		DisableCaller:     true,
	}
	zapLogger, _ := logCfg.Build()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	token := os.Getenv(envTelegramSecretToken)
	if token == "" {
		logger.Fatal("telegram secret token is empty")
	}

	var (
		offset int
		err    error
	)
	if os.Getenv(envTelegramOffset) != "" {
		offsetStr := os.Getenv(envTelegramOffset)
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "cannot parse offset"))
		}

		logger.Infof("Set offset to %d", offset)
	}

	logger.Info("Init application")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "cannot init bot"))
	}

	u := tgbotapi.NewUpdate(offset)
	u.Timeout = updateTimeout

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "cannot get updates"))
	}

	go handler(logger, bot, updates)
	logger.Error("Start listening messages")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	logger.Error("Gracefully shutdown.")

	bot.StopReceivingUpdates()
}

func handler(logger *zap.SugaredLogger, bot *tgbotapi.BotAPI, updates <-chan tgbotapi.Update) {
	var err error

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		logger.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		_, err = bot.Send(msg)
		if err != nil {
			logger.Error(errors.Wrap(err, "cannot send message"))
		}
	}
}
