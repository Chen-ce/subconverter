package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Chen-ce/subconverter/pkg/logger"
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
		client: &http.Client{Timeout: 60 * time.Second}, // 超时加大点，上传可能慢
	}, nil
}

// GenerateID 生成随机 ID
func (s *HFDatasetStorage) GenerateID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Save 保存配置到 HF Dataset (通过 commit add/update)
func (s *HFDatasetStorage) Save(cfg *SubscriptionConfig) error {
	cfg.UpdatedAt = time.Now()
	cfg.Version++

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	filePath := cfg.ID + ".json"
	base64Content := base64.StdEncoding.EncodeToString(data) // HF commit 需要 base64

	payload := map[string]interface{}{
		"title": "Update config " + cfg.ID, // commit title (required)
		"operations": []map[string]interface{}{
			{
				"op":       "add", // add 或 update 都用 add (如果存在会覆盖)
				"path":     filePath,
				"content":  base64Content,
				"encoding": "base64",
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://huggingface.co/api/datasets/%s/commit/main", s.repoID)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	logger.Debug("Committing to HF", "url", apiURL)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to commit to HF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		logger.Error("HF API error response", "status", resp.StatusCode, "body", string(respBody))
		return fmt.Errorf("HF commit error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Load 从 HF Dataset 加载配置 (raw URL, 完美)
func (s *HFDatasetStorage) Load(id string) (*SubscriptionConfig, error) {
	rawURL := fmt.Sprintf("https://huggingface.co/datasets/%s/raw/main/%s.json", s.repoID, id)

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.token) // 私有 repo 需要

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

// Delete 从 HF Dataset 删除配置 (通过 commit delete operation)
func (s *HFDatasetStorage) Delete(id string) error {
	filePath := id + ".json"

	payload := map[string]interface{}{
		"title": "Delete config " + id,
		"operations": []map[string]interface{}{
			{
				"op":   "delete",
				"path": filePath,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("HF commit error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// List 列出所有配置 (tree API 正确)
func (s *HFDatasetStorage) List() ([]*SubscriptionConfig, error) {
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
				continue // 跳过损坏的
			}
			configs = append(configs, cfg)
		}
	}

	return configs, nil
}

// Exists 检查是否存在
func (s *HFDatasetStorage) Exists(id string) bool {
	_, err := s.Load(id)
	return err == nil
}
