package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/canhlinh/hlsdl"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// func findVideoUrl(siteUrl string) chan string {
// 	c := colly.NewCollector()
// 	resp := make(chan string)

// 	c.OnRequest(func(r *colly.Request) {
// 		url := r.URL.String()
// 		if strings.HasPrefix(url, "https://rgevod.akamaized.net/vodedge") {
// 			fmt.Println("found: " + url)
// 			resp <- url
// 		} else {
// 			// fmt.Println("not matching: " + url)
// 			// rawBody, err := ioutil.ReadAll(r.Body)
// 			// if err != nil {
// 			// 	panic(err)
// 			// }
// 			// fmt.Println(rawBody)
// 		}
// 	})

// 	c.OnResponse(func(r *colly.Response) {
// 		fmt.Println("response")
// 		fmt.Println(string(r.Body))
// 	})

// 	c.Visit(siteUrl)
// 	return resp
// }

func findVideoUrl(siteUrl string) string {
	resp := make(chan string, 1)

	// Headless runs the browser on foreground, you can also use flag "-rod=show"
	// Devtools opens the tab in each new tab opened automatically
	// l := launcher.New().
	// 	NoSandbox(true).
	// 	Headless(false).
	// 	Devtools(true)

	// defer l.Cleanup() // remove launcher.FlagUserDataDir

	// curl := l.MustLaunch()

	browser := rod.New().
		// ControlURL(curl).
		// Trace(true).
		// SlowMotion(2 * time.Second).
		MustConnect()
	launcher.Open(browser.ServeMonitor(""))
	defer browser.MustClose()

	router := browser.HijackRequests()
	defer router.MustStop()

	router.MustAdd("https://trc.taboola.com*playlist.m3u8*", func(ctx *rod.Hijack) {
		// router.MustAdd("*", func(ctx *rod.Hijack) {
		url := ctx.Request.URL().Query().Get("data")
		_ = ctx.LoadResponse(http.DefaultClient, true)
		fmt.Println(url)
		resp <- url
	})

	go router.Run()
	browser.MustPage(siteUrl)

	urlObj := <-resp
	var taboolaUrlContainer taboolaReq
	if err := json.Unmarshal([]byte(urlObj), &taboolaUrlContainer); err != nil {
		panic(err)
	}

	taboolaUrl, err := url.Parse(taboolaUrlContainer.Url)
	if err != nil {
		panic(err)
	}

	vidUrl := taboolaUrl.Query().Get("videoUrl")

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

	return streamUrl
}

func download(url string) {
	fmt.Println(url)
	hlsDL := hlsdl.New(url, nil, fmt.Sprintf("download_%d", time.Now().Unix()), 64, true)
	filepath, err := hlsDL.Download()
	if err != nil {
		panic(err)
	}

	fmt.Println(filepath)
}

func main() {
	fmt.Println("start")
	// url := "https://sport5api.akamaized.net/Redirector/sport5/manifest/NM_VTR_TAK_RAINA_131122_2/HLS/playlist.m3u8"
	// url := "https://rgevod.akamaized.net/vodedge/_definst_/mp4:rge/bynet/sport5/sport5/PRV5/sc0YaB81bw/App/NM_VTR_TAK_RAINA_131122_2_1800.mp4/chunklist_b1800000.m3u8"
	// url := "https://rgevod.akamaized.net/vodedge/_definst_/mp4:rge/bynet/sport5/sport5/PRV5/sc0YaB81bw/App/NM_VTR_TAK_KASH_131122_2_1800.mp4/chunklist_b1800000.m3u8"
	// url := findVideoUrl("https://playbyplay.sport5.co.il/?GameID=125421&FLNum=16")
	p := NewParser()
	url, err := p.FindDownloadLink("https://playbyplay.sport5.co.il/?GameID=125423&FLNum=16")
	if err != nil {
		panic(err)
	}
	fmt.Println("found video url: " + url)
	download(url)
}
