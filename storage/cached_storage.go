package storage

import (
	"time"
	"github.com/Chen-ce/subconverter/pkg/cache"
	"github.com/Chen-ce/subconverter/pkg/logger"
)

// CachedStorage 带有内存缓存的存储包装器
type CachedStorage struct {
	base  Storage
	cache *cache.Cache
	ttl   time.Duration
}

// NewCachedStorage 创建一个新的带缓存存储
func NewCachedStorage(base Storage, ttl time.Duration) *CachedStorage {
	return &CachedStorage{
		base:  base,
		cache: cache.New(),
		ttl:   ttl,
	}
}

func (s *CachedStorage) GenerateID() (string, error) {
	return s.base.GenerateID()
}

func (s *CachedStorage) Save(cfg *SubscriptionConfig) error {
	err := s.base.Save(cfg)
	if err == nil {
		s.cache.Set(cfg.ID+"/core.json", cfg, s.ttl)
	}
	return err
}

func (s *CachedStorage) Load(id string) (*SubscriptionConfig, error) {
	key := id + "/core.json"
	if val, found := s.cache.Get(key); found {
		logger.Debug("Config loaded from cache", "id", id)
		return val.(*SubscriptionConfig), nil
	}
	
	cfg, err := s.base.Load(id)
	if err == nil {
		s.cache.Set(key, cfg, s.ttl)
	}
	return cfg, err
}

func (s *CachedStorage) Delete(id string) error {
	err := s.base.Delete(id)
	if err == nil {
		s.cache.Delete(id + "/core.json")
	}
	return err
}

func (s *CachedStorage) List() ([]*SubscriptionConfig, error) {
	return s.base.List()
}

func (s *CachedStorage) Exists(id string) bool {
	key := id + "/core.json"
	if _, found := s.cache.Get(key); found {
		return true
	}
	return s.base.Exists(id)
}

func (s *CachedStorage) SaveFile(id, filename string, data []byte) error {
	err := s.base.SaveFile(id, filename, data)
	if err == nil {
		s.cache.Set(id+"/"+filename, data, s.ttl)
	}
	return err
}

func (s *CachedStorage) LoadFile(id, filename string) ([]byte, error) {
	key := id + "/" + filename
	if val, found := s.cache.Get(key); found {
		logger.Debug("File loaded from cache", "path", key)
		return val.([]byte), nil
	}

	data, err := s.base.LoadFile(id, filename)
	if err == nil {
		s.cache.Set(key, data, s.ttl)
	}
	return data, err
}

func (s *CachedStorage) ClearCache(id string) error {
	err := s.base.ClearCache(id)
	if err == nil {
		// 删除内存中以此 ID 开头的所有缓存 (包括 core.json, cache_*, export_*)
		s.cache.DeletePrefix(id + "/")
	}
	return err
}
