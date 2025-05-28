package discord

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func PostEmbed(url string, embeds []Embed) (*http.Response, error) {
	data := discordParams{
		Embeds: embeds,
	}
	jsonData, err := json.Marshal(data)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	webhook := &http.Client{}
	response, err := webhook.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return response, err
}

type discordParams struct {
	Embeds []Embed `json:"embeds"`
}

type Embed struct {
	Title       string    `json:"title,omitempty"`
	Type        string    `json:"type,omitempty"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Color       int       `json:"color,omitempty"`
	Footer      Footer    `json:"footer"`
}

type Footer struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}
