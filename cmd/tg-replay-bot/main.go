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

func main() {
	zapLogger := prepareLogger()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	token := prepareToken(logger)
	offset := prepareOffset(logger)
	chatID := prepareChatID(logger)

	logger.Info("Init application")

	bot, updates := prepareBotWithUpdates(token, offset, logger)

	listenAndHandleBotUpdates(bot, updates, chatID)

	logger.Error("Bot started")

	waitSignalForGracefullyShutdown(bot, logger)

	logger.Error("Bot stopped")
}

func prepareLogger() *zap.Logger {
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

	return zapLogger
}

const envTelegramSecretToken = "TELEGRAM_SECRET_TOKEN"

func prepareToken(logger *zap.SugaredLogger) string {
	value := os.Getenv(envTelegramSecretToken)
	if value == "" {
		logger.Fatalf("%s is empty", envTelegramSecretToken)
	}

	return value
}

const envTelegramOffset = "TELEGRAM_OFFSET"

func prepareOffset(logger *zap.SugaredLogger) int {
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

	return offset
}

const envTelegramChatID = "TELEGRAM_CHAT_ID"

func prepareChatID(logger *zap.SugaredLogger) int64 {
	chat := os.Getenv(envTelegramChatID)
	if chat == "" {
		logger.Fatalf("%s is empty", envTelegramChatID)
	}

	chatID, err := strconv.ParseInt(chat, 10, 64)
	if err != nil {
		logger.Fatal(errors.Wrapf(err, "cannot parse %s", envTelegramChatID))
	}

	return chatID
}

const updateTimeout = 60 // seconds

func prepareBotWithUpdates(token string, offset int, logger *zap.SugaredLogger) (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel) {
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

	return bot, updates
}

func listenAndHandleBotUpdates(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, chatID int64) {
	errCh := make(chan error)

	botHandler := appbot.New(bot, errCh)
	chatHandler := chatbot.New(bot, errCh, chatID)

	r := router.New(botHandler, chatHandler)

	go r.Route(updates)
}

func waitSignalForGracefullyShutdown(bot *tgbotapi.BotAPI, logger *zap.SugaredLogger) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	logger.Error("Gracefully shutdown")

	bot.StopReceivingUpdates()
}
