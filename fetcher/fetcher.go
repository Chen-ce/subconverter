package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher 订阅获取器
type Fetcher struct {
	client    *http.Client
	userAgent string
}

// NewFetcher 创建订阅获取器
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		userAgent: "subconverter/1.0",
	}
}

// SetUserAgent 设置 User-Agent
func (f *Fetcher) SetUserAgent(ua string) {
	f.userAgent = ua
}

// Fetch 获取订阅内容
func (f *Fetcher) Fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", f.userAgent)
	
	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	
	return string(body), nil
}
