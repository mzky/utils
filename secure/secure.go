package secure

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SlowHTTPDefender 防御慢速HTTP攻击的中间件集合
type SlowHTTPDefender struct {
	// 连接限制
	maxConns    int
	connTracker map[string]int
	connMutex   sync.Mutex

	// 速率限制
	rateLimiters map[string]*rate.Limiter
	rateMutex    sync.Mutex
	ratePerSec   int
	burstSize    int

	// 请求体限制
	maxBodySize int64
	readTimeout time.Duration

	// 黑名单
	blacklist  map[string]time.Time
	blackMutex sync.Mutex
}

// NewSlowHTTPDefender 创建新的防御中间件实例
func NewSlowHTTPDefender(maxConns int, ratePerSec int, burstSize int, maxBodySize int64, readTimeout time.Duration) *SlowHTTPDefender {
	return &SlowHTTPDefender{
		maxConns:     maxConns,
		connTracker:  make(map[string]int),
		rateLimiters: make(map[string]*rate.Limiter),
		ratePerSec:   ratePerSec,
		burstSize:    burstSize,
		maxBodySize:  maxBodySize,
		readTimeout:  readTimeout,
		blacklist:    make(map[string]time.Time),
	}
}

// DefenseMiddleware 防御中间件 - 组合所有防御策略
func (s *SlowHTTPDefender) DefenseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 检查黑名单
		if s.isBlacklisted(clientIP) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 连接数限制
		if !s.allowConnection(clientIP) {
			s.addToBlacklist(clientIP)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		defer s.releaseConnection(clientIP)

		// 速率限制
		if !s.allowRate(clientIP) {
			s.addToBlacklist(clientIP)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		// 请求体读取限制
		c.Request.Body = &limitedReader{
			Reader:          c.Request.Body,
			maxSize:         s.maxBodySize,
			bodyReadTimer:   time.AfterFunc(s.readTimeout, func() { s.addToBlacklist(clientIP) }),
			bodyReadTimeout: s.readTimeout,
		}

		c.Next()
	}
}

// 检查IP是否在黑名单中（带过期检查）
func (s *SlowHTTPDefender) isBlacklisted(ip string) bool {
	s.blackMutex.Lock()
	defer s.blackMutex.Unlock()

	if t, exists := s.blacklist[ip]; exists {
		if time.Since(t) > 10*time.Minute { // 黑名单有效期10分钟
			delete(s.blacklist, ip)
			return false
		}
		return true
	}
	return false
}

// 添加IP到黑名单
func (s *SlowHTTPDefender) addToBlacklist(ip string) {
	s.blackMutex.Lock()
	s.blacklist[ip] = time.Now()
	s.blackMutex.Unlock()
}

// 检查请求速率
func (s *SlowHTTPDefender) allowRate(ip string) bool {
	s.rateMutex.Lock()
	limiter, exists := s.rateLimiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(s.ratePerSec), s.burstSize)
		s.rateLimiters[ip] = limiter
	}
	s.rateMutex.Unlock()

	return limiter.Allow()
}

// 改进后的allowConnection方法
func (s *SlowHTTPDefender) allowConnection(ip string) bool {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	// 获取当前IP的连接数
	count := s.connTracker[ip]

	// 检查是否超过最大连接限制
	if count >= s.maxConns {
		return false
	}

	// 增加该IP的连接计数
	s.connTracker[ip] = count + 1
	return true
}

// 改进后的releaseConnection方法
func (s *SlowHTTPDefender) releaseConnection(ip string) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	// 减少该IP的连接计数
	if count, exists := s.connTracker[ip]; exists {
		count--
		if count <= 0 {
			delete(s.connTracker, ip)
		} else {
			s.connTracker[ip] = count
		}
	}
}

// 改进后的limitedReader实现
type limitedReader struct {
	io.Reader
	maxSize         int64
	readSize        int64
	bodyReadTimer   *time.Timer
	mu              sync.Mutex
	closed          bool
	bodyReadTimeout time.Duration
}

func (l *limitedReader) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.closed {
		if !l.bodyReadTimer.Stop() {
			<-l.bodyReadTimer.C // drain timer if it has fired
		}
		l.closed = true
	}

	if closer, ok := l.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (l *limitedReader) Read(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 检查是否已关闭
	if l.closed {
		return 0, io.ErrClosedPipe
	}

	// 检查大小限制
	if l.readSize >= l.maxSize {
		return 0, io.EOF
	}

	// 计算剩余可用空间
	if max := l.maxSize - l.readSize; int64(len(p)) > max {
		p = p[:max]
	}

	// 重置定时器
	if l.bodyReadTimer != nil {
		if !l.bodyReadTimer.Stop() {
			<-l.bodyReadTimer.C // drain timer if it has fired
		}
		l.bodyReadTimer.Reset(l.bodyReadTimeout)
	}

	n, err = l.Reader.Read(p)
	if n > 0 {
		newSize := l.readSize + int64(n)
		if newSize < l.readSize || newSize > l.maxSize {
			// 整数溢出或超出限制
			if l.bodyReadTimer != nil {
				l.bodyReadTimer.Stop()
			}
			return 0, io.ErrShortBuffer
		}
		l.readSize = newSize
	}

	// 如果读取结束，停止定时器
	if err != nil && (err == io.EOF || errors.Is(err, context.DeadlineExceeded)) {
		if l.bodyReadTimer != nil {
			l.bodyReadTimer.Stop()
		}
	}

	return n, err
}
