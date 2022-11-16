package main

import (
	"errors"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramHandler interface {
	ListenAndServe() error
}

type TelegramConfig struct {
	Token   string
	IsDebug bool
}

var ErrEmptyToken = errors.New("empty token")

func NewTelegramHandler(scraper Scraper, cfg TelegramConfig) TelegramHandler {
	return &telegramHandler{
		scraper: scraper,
		cfg:     cfg,
	}
}

type telegramHandler struct {
	scraper Scraper
	cfg     TelegramConfig
}

func (t *telegramHandler) ListenAndServe() error {
	if t.cfg.Token == "" {
		return ErrEmptyToken
	}

	bot, err := tgbotapi.NewBotAPI(t.cfg.Token)
	if err != nil {
		return err
	}

	if t.cfg.IsDebug {
		bot.Debug = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil && IsStreamUrl(update.Message.Text) {
			plsWaitMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "זה ייקח כמה שניות ⌛️")
			plsWaitMsg.ReplyToMessageID = update.Message.MessageID

			bot.Send(plsWaitMsg)

			filePath, err := t.scraper.FindAndDownload(update.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "זה לא עבד 🙈\nאולי תנסו שוב עוד כמה שניות")
				msg.ReplyToMessageID = update.Message.MessageID

				bot.Send(msg)
				continue
			}

			videoMsg := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(filePath))
			videoMsg.SupportsStreaming = true
			videoMsg.ReplyToMessageID = update.Message.MessageID

			bot.Send(videoMsg)
			os.RemoveAll(filePath)
		}
	}

	return nil
}
