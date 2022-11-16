package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thatisuday/commando"
)

const flagNoCustomBrowser = "NO_CUSTOM_BROWSER"

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
		AddFlag("verbose", "display logs while serving the project", commando.Bool, false).
		AddFlag("custom-browser", "use a custom browser for scraping", commando.String, flagNoCustomBrowser).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			verbose, err := flags["verbose"].GetBool()
			if err != nil {
				log.WithError(err).Fatal("error parsing flag")
			}
			if verbose {
				fmt.Println("verbose")
				log.SetLevel(log.DebugLevel)
			}

			var customBrowser string
			if b, err := flags["custom-browser"].GetString(); err != nil {
				log.WithError(err).Fatal("error parsing flag")
			} else if b != flagNoCustomBrowser {
				customBrowser = b
				log.Debug("using custom browser, browser=", b)
			}

			s := NewScraper(
				NewParser(ParserConfiguration{
					DisplayBrowser: false,
					CustomBrowser:  customBrowser,
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
		AddFlag("verbose", "display logs while serving the project", commando.Bool, false).
		AddFlag("custom-browser", "use a custom browser for scraping", commando.String, flagNoCustomBrowser).
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
			if b, err := flags["custom-browser"].GetString(); err != nil {
				log.WithError(err).Fatal("error parsing flag")
			} else if b != flagNoCustomBrowser {
				customBrowser = b
				log.Debug("using custom browser, browser=", b)
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

			ctx, cancel := context.WithCancel(context.Background())
			go cancelOnSignal(cancel)
			done := make(chan bool, 1)
			go wait(ctx, done, 20*time.Second)
			if err := th.ListenAndServe(ctx, done); err != nil {
				log.WithError(err).Fatal("error listening")
			}
			log.Debug("telegram handler is done")
		})

	commando.Parse(nil)
}

func cancelOnSignal(cancel func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Debug("received signal, terminating... signal=", sig)
	cancel()
}

func wait(ctx context.Context, done <-chan bool, timeout time.Duration) {
	<-ctx.Done()
	select {
	case <-done:
		{
			log.Debug("terminated gracefully")
		}
	case <-time.After(timeout):
		{
			log.Fatal("could not terminate gracefully during timeout, timeout=", timeout)
		}
	}
}
