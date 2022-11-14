package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-rod/rod"
	log "github.com/sirupsen/logrus"
)

type Parser interface {
	FindDownloadLink(url string) (string, error)
}

func NewParser() Parser {
	return &parser{}
}

type parser struct{}

func (p *parser) FindDownloadLink(siteUrl string) (string, error) {
	urlObj := p.findTaboolaTraceUrl(siteUrl)
	log.Debug("found taboola trace url", urlObj)
	vidUrl := p.extractVideoInfoLink(urlObj)
	httpResp, err := http.Get(vidUrl)
	if err != nil {
		panic(err)
	}

	// body, err := ioutil.ReadAll(httpResp.Body)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(body))

	scanner := bufio.NewScanner(httpResp.Body)
	nextLine := false
	var streamUrl string
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		if nextLine {
			streamUrl = line
			nextLine = false
		}
		if line == "#EXT-X-STREAM-INF:BANDWIDTH=1800000" {
			nextLine = true
		}

	}

	fmt.Println("done")
	// <-time.After(3 * time.Minute)

	return streamUrl, nil
}

func (p *parser) findTaboolaTraceUrl(siteUrl string) string {
	resp := make(chan string, 1)
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	router := browser.HijackRequests()
	defer router.MustStop()

	router.MustAdd("https://trc.taboola.com*playlist.m3u8*", func(ctx *rod.Hijack) {
		url := ctx.Request.URL().Query().Get("data")
		_ = ctx.LoadResponse(http.DefaultClient, true)
		resp <- url
	})

	go router.Run()
	browser.MustPage(siteUrl)

	return <-resp
}

func (p *parser) extractVideoInfoLink(urlObj string) string {
	var taboolaUrlContainer taboolaReq
	if err := json.Unmarshal([]byte(urlObj), &taboolaUrlContainer); err != nil {
		panic(err)
	}

	taboolaUrl, err := url.Parse(taboolaUrlContainer.Url)
	if err != nil {
		panic(err)
	}

	return taboolaUrl.Query().Get("videoUrl")
}
