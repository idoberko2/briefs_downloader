package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

type TelegramHandler interface {
	ListenAndServe(ctx context.Context, done chan<- bool) error
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

func (t *telegramHandler) ListenAndServe(ctx context.Context, done chan<- bool) error {
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
	for {
		select {
		case update := <-updates:
			{
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

					log.Debug("removing file...", filePath)
					if err := os.RemoveAll(filepath.Dir(filePath)); err != nil {
						log.WithError(err).Error("error removing file")
					} else {
						log.Debug("removed file completed successfully", filePath)
					}
				}
			}
		case <-ctx.Done():
			{
				done <- true
				return nil
			}
		}
	}
}
