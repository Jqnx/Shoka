package nhentai

import (
	"Shoka/internal/config"
	"compress/gzip"
	"io"
	"net/http"
	"time"
)

func (s *Nhentai) Request(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", s.cfg.Sources.NHentai.UserAgent)
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Encoding", "gzip, deflate, br")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Cache-Control", "no-cache")

	request.AddCookie(&http.Cookie{
		Name:   "csrftoken",
		Value:  s.cfg.Sources.NHentai.CSRFToken,
		Domain: config.NHDomain,
		Path:   "/",
	})

	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		body, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return body, nil

	} else {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
}
