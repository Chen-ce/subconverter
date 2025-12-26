package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
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

// Save 保存配置到 HF Dataset
func (s *HFDatasetStorage) Save(cfg *SubscriptionConfig) error {
	cfg.UpdatedAt = time.Now()
	cfg.Version++

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return s.SaveFile(cfg.ID, "core.json", data)
}

// SaveFile 保存核心文件或缓存文件
func (s *HFDatasetStorage) SaveFile(id, filename string, data []byte) error {
	filePath := id + "/" + filename
	base64Content := base64.StdEncoding.EncodeToString(data)

	header := map[string]interface{}{
		"key": "header",
		"value": map[string]interface{}{
			"summary": "Update " + filePath,
		},
	}
	op := map[string]interface{}{
		"key": "file",
		"value": map[string]interface{}{
			"path":     filePath,
			"content":  base64Content,
			"encoding": "base64",
		},
	}

	return s.commit(header, op)
}

// commit 执行 NDJSON 提交
func (s *HFDatasetStorage) commit(header, operation map[string]interface{}) error {
	var buf bytes.Buffer
	hJson, _ := json.Marshal(header)
	buf.Write(hJson)
	buf.WriteByte('\n')
	oJson, _ := json.Marshal(operation)
	buf.Write(oJson)
	buf.WriteByte('\n')

	apiURL := fmt.Sprintf("https://huggingface.co/api/datasets/%s/commit/main", s.repoID)
	req, err := http.NewRequest("POST", apiURL, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HF commit error (%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// Load 从 HF Dataset 加载配置
func (s *HFDatasetStorage) Load(id string) (*SubscriptionConfig, error) {
	data, err := s.LoadFile(id, "core.json")
	if err != nil {
		return nil, err
	}

	var cfg SubscriptionConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}

// LoadFile 加载原始文件内容
func (s *HFDatasetStorage) LoadFile(id, filename string) ([]byte, error) {
	filePath := id + "/" + filename
	rawURL := fmt.Sprintf("https://huggingface.co/datasets/%s/raw/main/%s", s.repoID, filePath)

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HF API error (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// Delete 从 HF Dataset 删除配置 (通过 commit NDJSON 格式)
func (s *HFDatasetStorage) Delete(id string) error {
	header := map[string]interface{}{
		"key": "header",
		"value": map[string]interface{}{
			"summary": "Delete directory " + id,
		},
	}
	op := map[string]interface{}{
		"key": "deletedEntry",
		"value": map[string]interface{}{
			"path": id, // 删除整个目录
		},
	}

	return s.commit(header, op)
}
// ClearCache 清除该 ID 下的所有缓存文件 (保留 core.json)
func (s *HFDatasetStorage) ClearCache(id string) error {
	apiURL := fmt.Sprintf("https://huggingface.co/api/datasets/%s/tree/main/%s", s.repoID, id)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("list tree for clear cache failed: %d", resp.StatusCode)
	}

	var files []struct {
		Type string `json:"type"`
		Path string `json:"path"`
	}
	json.NewDecoder(resp.Body).Decode(&files)

	var operations []map[string]interface{}
	for _, f := range files {
		if f.Type == "file" && !strings.HasSuffix(f.Path, "/core.json") {
			operations = append(operations, map[string]interface{}{
				"key": "deletedEntry",
				"value": map[string]interface{}{
					"path": f.Path,
				},
			})
		}
	}

	if len(operations) == 0 {
		return nil
	}

	header := map[string]interface{}{
		"key": "header",
		"value": map[string]interface{}{
			"summary": "Clear cache for " + id,
		},
	}

	// 这里需要批量提交多行 NDJSON，我们复用之前的 commit 逻辑
	// 但之前的 commit 只支持单行 operation
	var buf bytes.Buffer
	hJson, _ := json.Marshal(header)
	buf.Write(hJson)
	buf.WriteByte('\n')
	for _, op := range operations {
		oJson, _ := json.Marshal(op)
		buf.Write(oJson)
		buf.WriteByte('\n')
	}

	apiURLCommit := fmt.Sprintf("https://huggingface.co/api/datasets/%s/commit/main", s.repoID)
	reqCommit, _ := http.NewRequest("POST", apiURLCommit, &buf)
	reqCommit.Header.Set("Authorization", "Bearer "+s.token)
	reqCommit.Header.Set("Content-Type", "application/x-ndjson")

	respCommit, err := s.client.Do(reqCommit)
	if err != nil {
		return err
	}
	defer respCommit.Body.Close()

	if respCommit.StatusCode != http.StatusOK && respCommit.StatusCode != http.StatusCreated {
		return fmt.Errorf("clear cache commit failed: %d", respCommit.StatusCode)
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
		if f.Type == "directory" {
			id := f.Path
			cfg, err := s.Load(id)
			if err != nil {
				continue
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
