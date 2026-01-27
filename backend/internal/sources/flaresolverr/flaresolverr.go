package flaresolverr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"Shoka/internal/config"

	"github.com/PuerkitoBio/goquery"
)

type FlaresolverrRequest struct {
	Solution Solution `json:"solution"`
	Status   string   `json:"status"`
	Message  string   `json:"message"`
	Start    int64    `json:"startTimestamp"`
	End      int64    `json:"endTimestamp"`
	Version  string   `json:"version"`
}

type Solution struct {
	URL       string      `json:"url"`
	Status    int64       `json:"status"`
	Headers   http.Header `json:"headers"`
	Response  string      `json:"response"`
	Cookies   []Cookie    `json:"cookies"`
	UserAgent string      `json:"userAgent"`
}

type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	Size     int     `json:"size"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	Session  bool    `json:"session"`
	SameSite string  `json:"sameSite"`
}

type FlaresolverrBody struct {
	Cmd string `json:"cmd"`
	URL string `json:"url"`
}

func Request(cfg *config.Config, url string) ([]byte, error) {
	var flaresolverrURL string
	if !strings.Contains(cfg.Sources.Flaresolverr.URL, "/v1") {
		flaresolverrURL = fmt.Sprintf("%s/v1", cfg.Sources.Flaresolverr.URL)
	} else {
		flaresolverrURL = cfg.Sources.Flaresolverr.URL
	}

	client := http.Client{}

	body := FlaresolverrBody{
		Cmd: "request.get",
		URL: url,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, flaresolverrURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var f FlaresolverrRequest
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(f.Solution.Response))
	if err != nil {
		return nil, err
	}

	var out []byte

	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		out, _ = io.ReadAll(strings.NewReader(s.Text()))
	})

	return out, nil
}
