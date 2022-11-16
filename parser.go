package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	log "github.com/sirupsen/logrus"
)

type Quality string

const QualityLow Quality = "#EXT-X-STREAM-INF:BANDWIDTH=700000"
const QualityMedium Quality = "#EXT-X-STREAM-INF:BANDWIDTH=1100000"
const QualityHigh Quality = "#EXT-X-STREAM-INF:BANDWIDTH=1800000"

type Parser interface {
	FindDownloadLink(url string) (string, error)
}

func NewParser(cfg ParserConfiguration) Parser {
	if cfg.Quality == "" {
		cfg.Quality = QualityMedium
	}

	return &parser{
		cfg: cfg,
	}
}

type ParserConfiguration struct {
	DisplayBrowser bool
	CustomBrowser  string
	Quality        Quality
}

type parser struct {
	cfg ParserConfiguration
}

func (p *parser) FindDownloadLink(siteUrl string) (string, error) {
	urlObj := p.findTaboolaTraceUrl(siteUrl)
	log.Debug("found taboola trace url", urlObj)

	vidUrl, err := p.extractVideoInfoLink(urlObj)
	if err != nil {
		return "", err
	}
	log.Debug("found video info url", vidUrl)

	streamUrl, err := p.findStreamUrl(vidUrl)
	if err != nil {
		return "", err
	}
	log.Debug("found stream url", streamUrl)

	return streamUrl, nil
}

func (p *parser) findTaboolaTraceUrl(siteUrl string) string {
	var browser *rod.Browser
	if p.cfg.DisplayBrowser {
		log.Debug("starting non-headless browser")
		l := launcher.New().
			NoSandbox(true).
			Headless(false).
			Devtools(true)
		defer l.Cleanup()
		curl := l.MustLaunch()
		browser = rod.New().
			ControlURL(curl).
			Trace(true)
	} else if p.cfg.CustomBrowser != "" {
		log.Debug("starting custom browser, browser=", p.cfg.CustomBrowser)
		l := launcher.New().Bin(p.cfg.CustomBrowser)
		defer l.Cleanup()
		browser = rod.New().ControlURL(l.MustLaunch())
	} else {
		log.Debug("starting default headless browser")
		browser = rod.New()
	}

	resp := make(chan string, 1)
	browser.MustConnect()

	if p.cfg.DisplayBrowser {
		launcher.Open(browser.ServeMonitor(""))
	}

	defer browser.MustClose()

	router := browser.HijackRequests()
	defer router.MustStop()

	router.MustAdd("https://trc.taboola.com*playlist.m3u8*", func(ctx *rod.Hijack) {
		log.Debug("found taboola request")
		url := ctx.Request.URL().Query().Get("data")
		resp <- url
	})

	go router.Run()
	log.Debug("navigating to page...")
	browser.MustPage(siteUrl)

	return <-resp
}

func (p *parser) extractVideoInfoLink(urlObj string) (string, error) {
	var taboolaUrlContainer taboolaReq
	if err := json.Unmarshal([]byte(urlObj), &taboolaUrlContainer); err != nil {
		return "", err
	}

	taboolaUrl, err := url.Parse(taboolaUrlContainer.Url)
	if err != nil {
		return "", err
	}

	return taboolaUrl.Query().Get("videoUrl"), nil
}

func (p *parser) findStreamUrl(vidUrl string) (string, error) {
	httpResp, err := http.Get(vidUrl)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(httpResp.Body)
	isNextLine := false
	for scanner.Scan() {
		line := scanner.Text()
		if isNextLine {
			return line, nil
		}
		if line == string(p.cfg.Quality) {
			isNextLine = true
		}
	}

	return "", errors.New("could not find video url")
}

type taboolaReq struct {
	Url string `json:"u"`
}
