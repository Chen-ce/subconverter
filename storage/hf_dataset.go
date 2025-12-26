package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HFDatasetStorage Hugging Face Dataset 存储实现
type HFDatasetStorage struct {
	repoID string
	token  string
	client *http.Client
}

// NewHFDatasetStorage 创建 HF Dataset 存储
func NewHFDatasetStorage(repoID, token string) (*HFDatasetStorage, error) {
	if repoID == "" || token == "" {
		return nil, fmt.Errorf("repo_id and token are required for HF storage")
	}
	return &HFDatasetStorage{
		repoID: repoID,
		token:  token,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// GenerateID 生成随机 ID
func (s *HFDatasetStorage) GenerateID() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Save 保存配置到 HF Dataset
func (s *HFDatasetStorage) Save(cfg *SubscriptionConfig) error {
	cfg.UpdatedAt = time.Now()
	cfg.Version++

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 使用 HF Hub API 上传文件 (PUT 方法)
	// https://huggingface.co/docs/hub/api#put-apireposrepo-typerepo-iduploadpath
	apiURL := fmt.Sprintf("https://huggingface.co/api/datasets/%s/upload/main/%s.json", s.repoID, cfg.ID)
	
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload to HF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HF API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// Load 从 HF Dataset 加载配置
func (s *HFDatasetStorage) Load(id string) (*SubscriptionConfig, error) {
	// 直接从 HF 原始数据链接下载 (或者通过 API)
	// https://huggingface.co/datasets/REPO_ID/raw/main/PATH
	rawURL := fmt.Sprintf("https://huggingface.co/datasets/%s/raw/main/%s.json", s.repoID, id)
	
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download from HF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("config not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HF API error (%d): %s", resp.StatusCode, string(body))
	}

	var cfg SubscriptionConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}

// Delete 从 HF Dataset 删除配置
func (s *HFDatasetStorage) Delete(id string) error {
	// HF API 目前没有直接删除单个文件的简单 REST API，通常需要提交 commit
	// 简单起见，这里可能需要调用 https://huggingface.co/api/datasets/REPO_ID/commit/main
	
	payload := map[string]interface{}{
		"summary": "Delete config " + id,
		"operations": []map[string]interface{}{
			{
				"operation": "delete",
				"path":      id + ".json",
			},
		},
	}
	
	body, _ := json.Marshal(payload)
	apiURL := fmt.Sprintf("https://huggingface.co/api/datasets/%s/commit/main", s.repoID)
	
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete from HF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HF API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// List 列出所有配置
func (s *HFDatasetStorage) List() ([]*SubscriptionConfig, error) {
	// 使用 HF API 列出文件
	// https://huggingface.co/api/datasets/REPO_ID/tree/main
	apiURL := fmt.Sprintf("https://huggingface.co/api/datasets/%s/tree/main", s.repoID)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list HF files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HF API error (%d): %s", resp.StatusCode, string(body))
	}

	var files []struct {
		Type string `json:"type"`
		Path string `json:"path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode HF response: %w", err)
	}

	var configs []*SubscriptionConfig
	for _, f := range files {
		if f.Type == "file" && strings.HasSuffix(f.Path, ".json") {
			id := strings.TrimSuffix(f.Path, ".json")
			cfg, err := s.Load(id)
			if err != nil {
				continue
			}
			configs = append(configs, cfg)
		}
	}

	return configs, nil
}

// Exists 检查配置是否存在
func (s *HFDatasetStorage) Exists(id string) bool {
	_, err := s.Load(id)
	return err == nil
}
