// middleware/cache.go
package ginplus

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kongbaiai2/yilang/goapp/internal/global"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	MaxAge      int                         // 缓存时间（秒），0 表示禁用
	CacheStore  CacheStore                  // 缓存存储后端
	GenerateKey func(c *gin.Context) string // 缓存键生成函数
	IgnorePaths []string                    // 不缓存的路径前缀列表
}

// CacheStore 缓存存储接口
type CacheStore interface {
	Get(key string) (data []byte, ok bool, expire time.Time)
	Set(key string, data []byte, expire time.Time)
	Delete(key string)
}

// memoryCache 内存缓存实现（线程安全）
type memoryCache struct {
	data sync.Map // map[string]cacheEntry
}

type cacheEntry struct {
	Value  []byte
	Expire time.Time
}

func NewMemoryCache() CacheStore {
	return &memoryCache{}
}

func (m *memoryCache) Get(key string) ([]byte, bool, time.Time) {
	if v, ok := m.data.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.Expire) {
			return entry.Value, true, entry.Expire
		}
		m.data.Delete(key) // 自动清理过期项
	}
	return nil, false, time.Time{}
}

func (m *memoryCache) Set(key string, data []byte, expire time.Time) {
	m.data.Store(key, cacheEntry{Value: data, Expire: expire})
}

func (m *memoryCache) Delete(key string) {
	m.data.Delete(key)
}

// responseWriter 用于捕获响应体和状态码
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	closed bool // 标记是否已写入

}

func (w *responseWriter) WriteHeader(statusCode int) {
	if !w.closed {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}
func (w *responseWriter) Write(data []byte) (int, error) {
	if w.closed {
		// 防止多次写入
		return len(data), nil
	}
	w.body.Write(data)
	n, err := w.ResponseWriter.Write(data)
	w.closed = true // 第一次写入后锁定
	return n, err
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Cache 缓存中间件
func Cache(cfg CacheConfig) gin.HandlerFunc {
	if cfg.CacheStore == nil {
		cfg.CacheStore = NewMemoryCache()
	}
	if cfg.GenerateKey == nil {
		cfg.GenerateKey = defaultCacheKey
	}
	if cfg.MaxAge <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if c.IsAborted() { // 如果前面中间件已 Abort，直接跳过
			return
		}
		// 跳过忽略路径
		for _, p := range cfg.IgnorePaths {
			if strings.HasPrefix(c.Request.URL.Path, p) {
				c.Next()
				return
			}
		}

		key := cfg.GenerateKey(c)
		global.LOG.Debugf(">>> Request Key: %s\n", key)

		now := time.Now()

		// 尝试从缓存读取
		if data, ok, _ := cfg.CacheStore.Get(key); ok {
			global.LOG.Debugf(">>> CACHE HIT!\n")
			global.LOG.Debugf(">>> Cached data: %s\n", string(data))

			hash := sha256.Sum256(data)
			etag := hex.EncodeToString(hash[:8])

			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", cfg.MaxAge))
			c.Header("ETag", `"`+etag+`"`)
			c.Header("X-Cache", "HIT")

			// 支持 If-None-Match
			if match := c.GetHeader("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					c.Status(http.StatusNotModified)
					return
				}
			}

			// c.Data(http.StatusOK, c.ContentType(), data)
			c.Data(http.StatusOK, "application/json; charset=utf-8", data)
			c.Abort()
			return
		}

		global.LOG.Debugf(">>> CACHE MISS!")

		// 未命中：捕获响应
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		c.Next()

		// 仅缓存成功的 JSON 响应
		if c.Writer.Status() == http.StatusOK && w.body.Len() > 0 {
			// 可选：检查 Content-Type 是否为 JSON（更安全）
			// 但为简化，此处直接缓存（由 IgnorePaths 控制范围）

			ttl := now.Add(time.Duration(cfg.MaxAge) * time.Second)
			cfg.CacheStore.Set(key, w.body.Bytes(), ttl)

			// 设置缓存头（仅在首次生成时）
			hash := sha256.Sum256(w.body.Bytes())
			etag := hex.EncodeToString(hash[:8])
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", cfg.MaxAge))
			c.Header("ETag", `"`+etag+`"`)
			c.Header("X-Cache", "MISS")
		}
	}
}

func defaultCacheKey(c *gin.Context) string {
	q := c.Request.URL.RawQuery
	if q != "" {
		q = "?" + q
	}
	return fmt.Sprintf("%s:%s%s", c.Request.Method, c.Request.URL.Path, q)
}
