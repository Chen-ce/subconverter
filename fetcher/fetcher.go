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
			Timeout: 60 * time.Second, // 增加超时时间以支持慢速订阅源
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		// 使用 Clash 的 User-Agent 以兼容更多订阅源
		userAgent: "clash-verge/v1.3.8",
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
		// 提供更友好的超时错误提示
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			return "", fmt.Errorf("订阅源 %s 响应超时（60秒），请检查订阅源是否可用或网络连接", url)
		}
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
