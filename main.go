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

			// "https://playbyplay.sport5.co.il/?GameID=125423&FLNum=16"
			s := DefaultScraper()
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
					Quality: QualityMedium,
				}),
				NewDownloader())
			token := args["token"].Value
			th := NewTelegramHandler(s, token)
			if err := th.ListenAndServe(); err != nil {
				log.WithError(err).Fatal("error listening")
			}
		})

	commando.Parse(nil)
}
