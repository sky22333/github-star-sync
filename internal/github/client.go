package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github-star-sync/internal/model"
)

const apiBase = "https://api.github.com"

// Client fetches public starred repositories.
type Client struct {
	httpClient *http.Client
	token      string
	userAgent  string
}

// New creates a GitHub API client. Token may be empty (60 req/h).
func New(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		token:      strings.TrimSpace(token),
		userAgent:  "github-star-sync",
	}
}

type apiRepo struct {
	FullName        string   `json:"full_name"`
	HTMLURL         string   `json:"html_url"`
	Description     string   `json:"description"`
	Language        string   `json:"language"`
	StargazersCount int      `json:"stargazers_count"`
	Topics          []string `json:"topics"`
	Archived        bool     `json:"archived"`
	Homepage        string   `json:"homepage"`
}

type starredItem struct {
	StarredAt time.Time `json:"starred_at"`
	Repo      apiRepo   `json:"repo"`
}

// ListStarred returns all public stars for username (paginated).
func (c *Client) ListStarred(ctx context.Context, username string) ([]model.Repo, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	var out []model.Repo
	page := 1
	for {
		url := fmt.Sprintf("%s/users/%s/starred?per_page=100&page=%d", apiBase, username, page)
		body, link, err := c.get(ctx, url)
		if err != nil {
			return nil, err
		}

		var items []starredItem
		if err := json.Unmarshal(body, &items); err != nil {
			// Fallback: plain repo array without star envelope.
			var repos []apiRepo
			if err2 := json.Unmarshal(body, &repos); err2 != nil {
				return nil, fmt.Errorf("decode starred page %d: %w", page, err)
			}
			for _, r := range repos {
				out = append(out, toModel(r, time.Time{}))
			}
		} else {
			for _, it := range items {
				out = append(out, toModel(it.Repo, it.StarredAt))
			}
		}

		if !hasNextPage(link) {
			break
		}
		page++
	}
	return out, nil
}

func toModel(r apiRepo, starredAt time.Time) model.Repo {
	return model.Repo{
		FullName:    r.FullName,
		HTMLURL:     r.HTMLURL,
		Description: strings.TrimSpace(r.Description),
		Language:    strings.TrimSpace(r.Language),
		Stars:       r.StargazersCount,
		Topics:      append([]string(nil), r.Topics...),
		Archived:    r.Archived,
		StarredAt:   starredAt,
		Homepage:    strings.TrimSpace(r.Homepage),
	}
}

func (c *Client) get(ctx context.Context, url string) ([]byte, string, error) {
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, "", err
		}
		req.Header.Set("Accept", "application/vnd.github.star+json")
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("request %s: %w", url, err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if readErr != nil {
			return nil, "", readErr
		}

		if resp.StatusCode == http.StatusOK {
			return body, resp.Header.Get("Link"), nil
		}

		if resp.StatusCode == http.StatusNotFound {
			return nil, "", fmt.Errorf("user not found or stars private")
		}

		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			if wait := retryAfter(resp); wait > 0 && attempt < 3 {
				select {
				case <-ctx.Done():
					return nil, "", ctx.Err()
				case <-time.After(wait):
					continue
				}
			}
			return nil, "", fmt.Errorf("rate limited (%s): %s", resp.Status, strings.TrimSpace(string(body)))
		}

		return nil, "", fmt.Errorf("GET %s: %s: %s", url, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil, "", fmt.Errorf("GET %s: exceeded retries", url)
}

func retryAfter(resp *http.Response) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil && sec > 0 {
			return time.Duration(sec) * time.Second
		}
	}
	if v := resp.Header.Get("X-RateLimit-Remaining"); v == "0" {
		if reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
			d := time.Until(time.Unix(reset, 0))
			if d > 0 && d < time.Hour {
				return d + time.Second
			}
		}
	}
	return 2 * time.Second
}

func hasNextPage(link string) bool {
	// Link: <url>; rel="next", <url>; rel="last"
	for _, part := range strings.Split(link, ",") {
		if strings.Contains(part, `rel="next"`) {
			return true
		}
	}
	return false
}
