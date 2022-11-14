package main

type Scraper interface {
	FindAndDownload(url string) (string, error)
}

func NewScraper(p Parser, d Downloader) Scraper {
	return &scraper{
		p: p,
		d: d,
	}
}

func DefaultScraper() Scraper {
	return &scraper{
		p: NewParser(),
		d: NewDownloader(),
	}
}

type scraper struct {
	p Parser
	d Downloader
}

func (s *scraper) FindAndDownload(url string) (string, error) {
	streamUrl, err := s.p.FindDownloadLink(url)
	if err != nil {
		return "", err
	}

	filePath, err := s.d.Download(streamUrl)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
