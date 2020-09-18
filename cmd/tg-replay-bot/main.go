package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	appbot "github.com/kulaginds/tg-replay-bot/internal/app/bot"
	chatbot "github.com/kulaginds/tg-replay-bot/internal/app/chat"
	"github.com/kulaginds/tg-replay-bot/internal/pkg/router"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envTelegramSecretToken = "TELEGRAM_SECRET_TOKEN"
	envTelegramOffset      = "TELEGRAM_OFFSET"
	envTelegramChatID      = "TELEGRAM_CHAT_ID"

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

	chat := os.Getenv(envTelegramChatID)
	if chat == "" {
		logger.Fatal("telegram chat id is empty")
	}
	chatID, err := strconv.ParseInt(chat, 10, 64)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "cannot parse chat id"))
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

	errCh := make(chan error)

	botHandler := appbot.New(bot, errCh)
	chatHandler := chatbot.New(bot, errCh, chatID)

	r := router.New(botHandler, chatHandler)

	go r.Route(updates)
	logger.Error("Start listening messages")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	logger.Error("Gracefully shutdown.")

	bot.StopReceivingUpdates()
}
