package main

import (
	"fmt"
	"time"

	"github.com/canhlinh/hlsdl"
)

type Downloader interface {
	Download(streamUrl string) (string, error)
}

func NewDownloader() Downloader {
	return &downloader{}
}

type downloader struct{}

func (d *downloader) Download(streamUrl string) (string, error) {
	hlsDL := hlsdl.New(streamUrl, nil, fmt.Sprintf("download_%d", time.Now().Unix()), 64, true)
	filepath, err := hlsDL.Download()
	if err != nil {
		return "", err
	}

	return filepath, nil
}
