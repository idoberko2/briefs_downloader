package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/thatisuday/commando"
)

func main() {
	commando.
		SetExecutableName("dnldr").
		SetVersion("v1.0.0").
		SetDescription("This CLI tool scrapes and downloads short sports streams")

	commando.
		Register("download").
		SetDescription("This command scrapes and downloads a stream").
		SetShortDescription("download a stream").
		AddArgument("url", "url of the webpage", "").
		AddFlag("custom-browser", "use a custom browser for scraping", commando.String, nil).
		AddFlag("verbose", "display logs while serving the project", commando.Bool, nil).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			verbose, err := flags["verbose"].GetBool()
			if err != nil {
				log.WithError(err).Fatal("error parsing flag")
			}
			if verbose {
				fmt.Println("verbose")
				log.SetLevel(log.DebugLevel)
			}

			s := NewScraper(
				NewParser(ParserConfiguration{
					DisplayBrowser: false,
					CustomBrowser:  args["custom-browser"].Value,
					Quality:        QualityHigh,
				}),
				NewDownloader(),
			)
			url := args["url"].Value
			filePath, err := s.FindAndDownload(url)
			if err != nil {
				log.WithError(err).Fatal("failed to scrape link")
			}

			log.Info("file is downloaded to: " + filePath)
		})

	commando.
		Register("listen").
		SetDescription("This command listens via telegram and downloads a stream on demand").
		SetShortDescription("telegram listener").
		AddArgument("token", "telegram access token", "").
		AddFlag("custom-browser", "use a custom browser for scraping", commando.String, nil).
		AddFlag("verbose", "display logs while serving the project", commando.Bool, nil).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			verbose, err := flags["verbose"].GetBool()
			if err != nil {
				log.WithError(err).Fatal("error parsing flag")
			}
			if verbose {
				log.SetLevel(log.DebugLevel)
				log.Debug("set DEBUG verbosity")
			}
			var customBrowser string
			if cb, ok := flags["custom-browser"]; ok {
				if b, err := cb.GetString(); err != nil {
					log.WithError(err).Fatal("error parsing flag")
				} else {
					customBrowser = b
					log.Debug("using custom browser, browser=", b)
				}
			}

			s := NewScraper(
				NewParser(ParserConfiguration{
					Quality:        QualityMedium,
					DisplayBrowser: false,
					CustomBrowser:  customBrowser,
				}),
				NewDownloader())
			token := args["token"].Value
			th := NewTelegramHandler(s, TelegramConfig{
				Token:   token,
				IsDebug: verbose,
			})
			if err := th.ListenAndServe(); err != nil {
				log.WithError(err).Fatal("error listening")
			}
		})

	commando.Parse(nil)
}
