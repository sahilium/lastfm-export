package lastfm

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL       = "https://ws.audioscrobbler.com/2.0/"
	pageSize      = 200
	requestDelay  = 1 * time.Second
	maxRetries    = 8
	initialBackoff = 2 * time.Second
)

type Client struct {
	apiKey   string
	username string
	http     *http.Client
}

func NewClient(apiKey, username string) *Client {
	return &Client{
		apiKey:   apiKey,
		username: username,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type RecentTracksResponse struct {
	RecentTracks struct {
		Track []Track `json:"track"`
		Attr  struct {
			TotalPages string `json:"totalPages"`
			Total      string `json:"total"`
			Page       string `json:"page"`
		} `json:"@attr"`
	} `json:"recenttracks"`
}

type Track struct {
	Artist struct {
		Text string `json:"#text"`
		MBID string `json:"mbid"`
		Name string `json:"name"`
	} `json:"artist"`
	Album struct {
		Text string `json:"#text"`
		MBID string `json:"mbid"`
	} `json:"album"`
	Name   string `json:"name"`
	MBID   string `json:"mbid"`
	Date   *struct {
		UTS string `json:"uts"`
	} `json:"date"`
	Loved string `json:"loved"`
	URL   string `json:"url"`
}

type PageResult struct {
	Tracks     []Track
	TotalPages int
	Total      int
	Page       int
}

func (c *Client) GetRecentTracks(page int) (*PageResult, error) {
	params := url.Values{}
	params.Set("method", "user.getRecentTracks")
	params.Set("user", c.username)
	params.Set("api_key", c.apiKey)
	params.Set("format", "json")
	params.Set("limit", fmt.Sprintf("%d", pageSize))
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("extended", "1")

	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		reqURL := baseURL + "?" + params.Encode()

		resp, err := c.http.Get(reqURL)
		if err != nil {
			slog.Warn("last.fm request failed", "attempt", attempt+1, "err", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			slog.Warn("failed to read response", "attempt", attempt+1, "err", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			slog.Warn("rate limited by last.fm", "attempt", attempt+1, "sleep", backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			slog.Warn("unexpected status from last.fm", "status", resp.StatusCode, "attempt", attempt+1)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		var apiResp struct {
			RecentTracks struct {
				Track []Track `json:"track"`
				Attr  struct {
					TotalPages string `json:"totalPages"`
					Total      string `json:"total"`
					Page       string `json:"page"`
				} `json:"@attr"`
			} `json:"recenttracks"`
			Error   int    `json:"error"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			slog.Warn("failed to decode response", "attempt", attempt+1, "err", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if apiResp.Error != 0 {
			if apiResp.Error == 29 {
				slog.Warn("rate limited by last.fm api", "attempt", attempt+1, "sleep", backoff)
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return nil, fmt.Errorf("last.fm api error %d: %s", apiResp.Error, apiResp.Message)
		}

		totalPages := 1
		fmt.Sscanf(apiResp.RecentTracks.Attr.TotalPages, "%d", &totalPages)
		total := 0
		fmt.Sscanf(apiResp.RecentTracks.Attr.Total, "%d", &total)
		pageNum := 1
		fmt.Sscanf(apiResp.RecentTracks.Attr.Page, "%d", &pageNum)

		return &PageResult{
			Tracks:     apiResp.RecentTracks.Track,
			TotalPages: totalPages,
			Total:      total,
			Page:       pageNum,
		}, nil
	}

	return nil, fmt.Errorf("failed after %d retries", maxRetries)
}

func (c *Client) RateLimit() {
	time.Sleep(requestDelay)
}
