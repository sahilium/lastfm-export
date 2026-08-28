package musicbrainz

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
	baseURL       = "https://musicbrainz.org/ws/2"
	requestDelay  = 1 * time.Second
	maxRetries    = 5
	initialBackoff = 2 * time.Second
)

type Client struct {
	userAgent string
	http      *http.Client
}

func NewClient(userAgent string) *Client {
	return &Client{
		userAgent: userAgent,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ArtistSearchResult struct {
	Artists []ArtistMB `json:"artists"`
}

type ArtistMB struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SortName    string `json:"sort-name"`
	Disambiguation string `json:"disambiguation"`
	Country     string `json:"country"`
	Type        string `json:"type"`
	BeginArea   *struct {
		Name string `json:"name"`
	} `json:"begin-area"`
	EndArea *struct {
		Name string `json:"name"`
	} `json:"end-area"`
	LifeSpan *struct {
		Begin string `json:"begin"`
		End   string `json:"end"`
	} `json:"life-span"`
	Tags []struct {
		Count int    `json:"count"`
		Name  string `json:"name"`
	} `json:"tags"`
}

type ReleaseSearchResult struct {
	Releases []ReleaseMB `json:"releases"`
}

type ReleaseMB struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ArtistCredit []struct {
		Name string `json:"name"`
	} `json:"artist-credit"`
	Date        string `json:"date"`
	Country     string `json:"country"`
	Status      string `json:"status"`
	ReleaseGroup *struct {
		ID   string `json:"id"`
		Type string `json:"primary-type"`
	} `json:"release-group"`
}

type RecordingSearchResult struct {
	Recordings []RecordingMB `json:"recordings"`
}

type RecordingMB struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Length      int    `json:"length"`
	ArtistCredit []struct {
		Name string `json:"name"`
	} `json:"artist-credit"`
	Releases []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"releases"`
}

func (c *Client) doRequest(url string) ([]byte, error) {
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			slog.Warn("musicbrainz request failed", "attempt", attempt+1, "err", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			slog.Warn("failed to read musicbrainz response", "attempt", attempt+1, "err", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			slog.Warn("rate limited by musicbrainz", "attempt", attempt+1, "sleep", backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			slog.Warn("unexpected status from musicbrainz", "status", resp.StatusCode, "attempt", attempt+1)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("failed after %d retries", maxRetries)
}

func (c *Client) SearchArtist(query string) (*ArtistSearchResult, error) {
	url := fmt.Sprintf("%s/artist/?query=%s&fmt=json&limit=5", baseURL, 	url.QueryEscape(query))

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var result ArtistSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode artist search: %w", err)
	}

	return &result, nil
}

func (c *Client) GetArtist(mbid string) (*ArtistMB, error) {
	url := fmt.Sprintf("%s/artist/%s?fmt=json", baseURL, mbid)

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var artist ArtistMB
	if err := json.Unmarshal(body, &artist); err != nil {
		return nil, fmt.Errorf("decode artist: %w", err)
	}

	return &artist, nil
}

func (c *Client) SearchRelease(query string) (*ReleaseSearchResult, error) {
	url := fmt.Sprintf("%s/release/?query=%s&fmt=json&limit=5", baseURL, 	url.QueryEscape(query))

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var result ReleaseSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode release search: %w", err)
	}

	return &result, nil
}

func (c *Client) SearchRecording(query string) (*RecordingSearchResult, error) {
	url := fmt.Sprintf("%s/recording/?query=%s&fmt=json&limit=5", baseURL, 	url.QueryEscape(query))

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var result RecordingSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode recording search: %w", err)
	}

	return &result, nil
}

func (c *Client) RateLimit() {
	time.Sleep(requestDelay)
}
