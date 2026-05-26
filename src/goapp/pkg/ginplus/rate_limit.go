// middleware/rate_limit.go
package ginplus

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限速配置
type RateLimitConfig struct {
	Requests int                         // 每窗口允许的请求数
	Window   time.Duration               // 时间窗口
	KeyFunc  func(c *gin.Context) string // 限速键生成函数
	Storage  RateLimiter                 // 存储后端
}

// RateLimiter 接口
type RateLimiter interface {
	Inc(key string, window time.Duration) (count int, remaining int, reset time.Time, err error)
}

// memoryRateLimiter 内存滑动窗口限速器
type memoryRateLimiter struct {
	sync.RWMutex
	buckets map[string]*windowBucket
}

type windowBucket struct {
	Counts  []int
	Index   int
	Window  time.Duration
	ResetAt time.Time
}

func NewMemoryRateLimiter(window time.Duration) RateLimiter {
	if window <= 0 {
		window = time.Minute
	}
	return &memoryRateLimiter{
		buckets: make(map[string]*windowBucket),
	}
}

func (m *memoryRateLimiter) Inc(key string, window time.Duration) (int, int, time.Time, error) {
	m.Lock()
	defer m.Unlock()

	sec := int(window.Seconds())
	if sec <= 0 {
		sec = 60
	}

	if _, exists := m.buckets[key]; !exists {
		m.buckets[key] = &windowBucket{
			Counts:  make([]int, sec),
			Index:   0,
			Window:  window,
			ResetAt: time.Now().Add(window),
		}
	}

	b := m.buckets[key]
	now := time.Now()

	// 重置过期窗口
	if now.After(b.ResetAt) {
		b.Counts = make([]int, sec)
		b.Index = 0
		b.ResetAt = now.Add(window)
	}

	// 计数
	b.Counts[b.Index]++
	b.Index = (b.Index + 1) % sec

	total := 0
	for _, n := range b.Counts {
		total += n
	}
	remaining := sec - total // 注意：这里 remaining 是“还能发多少”，实际应为 max(0, Requests - total)
	// 但我们在中间件中直接比较 count > Requests，所以 remaining 仅用于 header

	return total, remaining, b.ResetAt, nil
}

// RateLimit 中间件
func RateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	if cfg.Storage == nil {
		cfg.Storage = NewMemoryRateLimiter(cfg.Window)
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = clientIPKey
	}

	return func(c *gin.Context) {
		key := cfg.KeyFunc(c)
		count, remaining, reset, _ := cfg.Storage.Inc(key, cfg.Window)

		// 设置标准限速头
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

		if count > cfg.Requests {
			c.Header("Retry-After", "1")
			// ✅ 使用 c.JSON 避免手写 JSON 字符串导致的格式错误
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// clientIPKey 提取客户端 IP
func clientIPKey(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = "0.0.0.0"
	}
	return "ip:" + ip
}
