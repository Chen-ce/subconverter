package fetcher

import (
	"fmt"
	"github.com/Chen-ce/subconverter/pkg/logger"
	"io"
	"net/http"
	"strings"
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
	
	// 根据 URL 参数智能选择 User-Agent
	userAgent := f.detectUserAgent(url)
	req.Header.Set("User-Agent", userAgent)
	
	logger.Debug("Fetching subscription", "url", url, "ua", userAgent)
	
	resp, err := f.client.Do(req)
	if err != nil {
		// 提供更友好的超时错误提示
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			logger.Error("Subscription fetch timeout", "url", url, "timeout", "60s")
			return "", fmt.Errorf("订阅源 %s 响应超时（60秒），请检查订阅源是否可用或网络连接", url)
		}
		logger.Error("Subscription fetch failed", "url", url, "error", err)
		return "", fmt.Errorf("failed to fetch subscription from %s: %w", url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		logger.Error("Subscription fetch unexpected status", "url", url, "status", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body from %s: %w", url, err)
	}
	
	return string(body), nil
}

// detectUserAgent 根据 URL 参数智能检测应该使用的 User-Agent
func (f *Fetcher) detectUserAgent(url string) string {
	urlLower := strings.ToLower(url)
	
	// 检查 URL 中的客户端标识参数
	if strings.Contains(urlLower, "flag=clash") || strings.Contains(urlLower, "target=clash") {
		return "clash-verge/v1.3.8"
	}
	if strings.Contains(urlLower, "flag=surge") || strings.Contains(urlLower, "target=surge") {
		return "Surge/5.0.0"
	}
	if strings.Contains(urlLower, "flag=v2ray") || strings.Contains(urlLower, "target=v2ray") {
		return "v2rayN/6.23"
	}
	if strings.Contains(urlLower, "flag=singbox") || strings.Contains(urlLower, "target=singbox") {
		return "sing-box/1.8.0"
	}
	if strings.Contains(urlLower, "flag=quantumult") || strings.Contains(urlLower, "target=quanx") {
		return "Quantumult%20X/1.4.1"
	}
	if strings.Contains(urlLower, "flag=loon") || strings.Contains(urlLower, "target=loon") {
		return "Loon/3.2.0"
	}
	
	// 默认使用 Clash User-Agent（最通用）
	return "clash-verge/v1.3.8"
}
