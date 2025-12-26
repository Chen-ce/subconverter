package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// URLHash 生成 URL 的稳定 SHA-256 哈希值
func URLHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:16]) // 取前 16 位通常足够且长度友好
}
