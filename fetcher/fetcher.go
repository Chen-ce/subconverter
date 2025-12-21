package fetcher

import (
	"context"
	"fmt"
	"io"
	"net"
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
	// 使用自定义 DNS 解析器（Cloudflare DNS）
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Second * 10,
				}
				// 使用 Cloudflare DNS (1.1.1.1) 和 Google DNS (8.8.8.8) 作为备份
				return d.DialContext(ctx, network, "1.1.1.1:53")
			},
		},
	}

	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Fetcher{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
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
		return "", fmt.Errorf("failed to create request for %s: %w", url, err)
	}
	
	req.Header.Set("User-Agent", f.userAgent)
	
	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subscription from %s: %w", url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %w", url, err)
	}
	
	return string(body), nil
}
