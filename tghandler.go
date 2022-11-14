package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramHandler interface {
	ListenAndServe() error
}

func NewTelegramHandler(scraper Scraper, token string) TelegramHandler {
	return &telegramHandler{
		scraper: scraper,
		token:   token,
	}
}

type telegramHandler struct {
	scraper Scraper
	token   string
}

func (t *telegramHandler) ListenAndServe() error {
	bot, err := tgbotapi.NewBotAPI(t.token)
	if err != nil {
		return err
	}

	bot.Debug = true

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
		}
	}

	return nil
}
