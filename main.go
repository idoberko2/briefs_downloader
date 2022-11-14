package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func main() {
	// url := "https://sport5api.akamaized.net/Redirector/sport5/manifest/NM_VTR_TAK_RAINA_131122_2/HLS/playlist.m3u8"
	// url := "https://rgevod.akamaized.net/vodedge/_definst_/mp4:rge/bynet/sport5/sport5/PRV5/sc0YaB81bw/App/NM_VTR_TAK_RAINA_131122_2_1800.mp4/chunklist_b1800000.m3u8"
	// url := "https://rgevod.akamaized.net/vodedge/_definst_/mp4:rge/bynet/sport5/sport5/PRV5/sc0YaB81bw/App/NM_VTR_TAK_KASH_131122_2_1800.mp4/chunklist_b1800000.m3u8"
	// url := findVideoUrl("https://playbyplay.sport5.co.il/?GameID=125421&FLNum=16")
	s := DefaultScraper()
	filePath, err := s.FindAndDownload("https://playbyplay.sport5.co.il/?GameID=125423&FLNum=16")
	if err != nil {
		log.WithError(err).Fatal("failed to scrape link")
	}

	fmt.Println("file is downloaded to: " + filePath)
}
