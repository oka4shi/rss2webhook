package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/mmcdole/gofeed"
	"github.com/oka4shi/rss2webhook/pkg/discord"
)

type Config struct {
	Items []Item `yaml:"items"`
}

type Item struct {
	Target       string    `yaml:"target"`
	WebhookURL   string    `yaml:"webhook_url"`
	Color        string    `yaml:"color"`
	Interval     int       `yaml:"interval"`
	LastAccessed time.Time `yaml:"last_accessed"`
	Errors       []string  `yaml:"errors"`
}

func main() {
	configPath := os.Getenv("R2W_CONFIG")
	if configPath == "" {
		configPath = "./config.yml"
	}

	configStr, err := readFile(configPath)
	if err != nil {
		log.Panicln(err)
	}

	var config Config
	err = yaml.Unmarshal(configStr, &config)
	if err != nil {
		log.Panicln(err)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for i, item := range config.Items {
		wg.Add(1)
		go func() {
			defer wg.Done()

			interval := time.Duration(item.Interval) * time.Minute
			if item.LastAccessed.Add(interval).After(time.Now()) {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			fp := gofeed.NewParser()
			feed, err := fp.ParseURLWithContext(item.Target, ctx)
			if err != nil {
				log.Printf("Error fetching feed %s: %v", item.Target, err)
				return
			}

			color := parseColor(item.Color)

			for j := range feed.Items {
				feedItem := feed.Items[len(feed.Items)-j-1]

				errIdx := slices.Index(item.Errors, feedItem.GUID)
				if errIdx != -1 {
					mu.Lock()
					config.Items[i].Errors = slices.Delete(config.Items[i].Errors, errIdx, errIdx+1)
					mu.Unlock()
				}

				if item.LastAccessed.After(*feedItem.PublishedParsed) && errIdx == -1 {
					log.Printf("Skipping item %s", feedItem.GUID)
					continue
				}

				embed := []discord.Embed{
					{
						Title:       html.UnescapeString(feedItem.Title),
						Type:        "rich",
						Description: "",
						URL:         feedItem.Link,
						Timestamp:   *feedItem.PublishedParsed,
						Color:       color,
						Footer: discord.Footer{
							Text: feed.Title,
						},
					},
				}

				resp, err := discord.PostEmbed(item.WebhookURL, embed)
				errmsg := ""
				if err != nil {
					errmsg = fmt.Sprintf("Error sending webhook for %s: %v", item.Target, err)
				} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
					errmsg = fmt.Sprintf("Error response from webhook for %s: %d %s", item.Target, resp.StatusCode, resp.Status)
				}
				if errmsg != "" {
					log.Println(errmsg)
					mu.Lock()
					config.Items[i].Errors = append(config.Items[i].Errors, feedItem.GUID)
					mu.Unlock()
					continue
				}
				log.Printf("Sent webhook: %s", feedItem.GUID)
			}
			mu.Lock()
			config.Items[i].LastAccessed = time.Now()
			mu.Unlock()
		}()

		wg.Wait()
	}

	if err := writeConfigFile(configPath, config); err != nil {
		log.Panicln(err)
	}

}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	str, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return str, nil
}

func writeConfigFile(path string, config Config) error {
	configData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	_, err = f.Write(configData)
	if err != nil {
		return err
	}
	return nil
}

func parseColor(hex string) int {
	i, err := strconv.ParseUint(strings.Replace(hex, "#", "", -1), 16, 24)
	if err != nil {
		log.Printf("Error parsing color %s: %v", hex, err)
		return 0x000000
	}
	return int(i)
}
