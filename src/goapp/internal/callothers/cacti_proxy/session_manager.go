// session_manager.go
package cacti_proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
)

// SessionManager 管理 Cacti v0.8 的会话生命周期（登录、保活、失效检测）
type SessionManager struct {
	baseURL  string
	username string
	password string
	client   *http.Client

	// 🔐 状态与锁
	loginMu    sync.RWMutex
	isLoggedIn bool

	// 📈 统计（可选上报）
	loginCount uint64
}

// NewSessionManager 创建新的 SessionManager 实例
func NewSessionManager(baseURL, username, password string, client *http.Client) *SessionManager {
	return &SessionManager{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		username:   username,
		password:   password,
		client:     client,
		isLoggedIn: false,
	}
}

func (sm *SessionManager) FlushLogin() {
	sm.isLoggedIn = false
}

// IsAlive 检查当前 session 是否仍有效（只读，不触发登录）
// 返回 true 表示已登录且可访问 graph_view.php；false 表示需重登录
func (sm *SessionManager) IsAlive() bool {
	sm.loginMu.RLock()
	if !sm.isLoggedIn {
		sm.loginMu.RUnlock()
		return false
	}
	sm.loginMu.RUnlock()

	testURL := sm.baseURL + "/graph_view.php"
	resp, err := sm.client.Get(testURL)
	if err != nil {
		global.LOG.Warnf("Session health check failed (GET %s): %v", testURL, err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// 检测是否被重定向到登录页或返回登录 HTML
	isLoginPage := bytes.Contains(body, []byte("登录到Cacti")) ||
		bytes.Contains(body, []byte("<title>登录")) ||
		(resp.StatusCode == 302 && strings.Contains(resp.Header.Get("Location"), "login.php"))

	if isLoginPage {
		global.LOG.Debug("Session expired or invalid (detected login page)")
		return false
	}
	return true
}

// ForceLogin 强制执行一次完整登录（提取 token → POST → 验证）
// 成功返回 nil；失败返回 error（含重试信息）
func (sm *SessionManager) ForceLogin() error {
	const maxRetries = 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		// 💡 指数退避：第 0 次不等待，第 1 次等 1s，第 2 次等 2s，第 3 次等 4s...
		if i > 0 {
			wait := time.Second << uint(i-1)
			global.LOG.Debugf("Retry login #%d after %v...", i+1, wait)
			time.Sleep(wait)
		}

		err := sm.doLoginOnce()
		if err == nil {
			sm.isLoggedIn = true
			sm.loginCount++
			global.LOG.Infof("✅ Session login successful (attempt #%d)", i+1)
			return nil
		}

		lastErr = err
		global.LOG.Warnf("Login attempt #%d failed: %v", i+1, err)
	}

	return fmt.Errorf("login failed after %d attempts: %w", maxRetries, lastErr)
}

// doLoginOnce 执行单次登录（无重试逻辑）
func (sm *SessionManager) doLoginOnce() error {
	loginURL := sm.baseURL + "/index.php"

	// 1️⃣ 提取 CSRF token
	token, err := sm.extractCSRFToken(loginURL)
	if err != nil {
		return fmt.Errorf("extract CSRF token: %w", err)
	}

	// 2️⃣ 发起登录请求
	loginData := url.Values{}
	loginData.Set("__csrf_magic", token)
	loginData.Set("action", "login")
	loginData.Set("login_username", sm.username)
	loginData.Set("login_password", sm.password)
	loginData.Set("remember", "on")

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(loginData.Encode()))
	if err != nil {
		return fmt.Errorf("build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginURL)

	resp, err := sm.client.Do(req)
	if err != nil {
		return fmt.Errorf("send login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("login response status %d", resp.StatusCode)
	}

	// 3️⃣ 验证登录结果：访问 graph_view.php
	if !sm.isLoginValid() {
		return fmt.Errorf("login succeeded but session not valid (graph_view.php returns login page)")
	}

	return nil
}

// extractCSRFToken 从 index.php 页面提取 __csrf_magic 值
func (sm *SessionManager) extractCSRFToken(loginURL string) (string, error) {
	resp, err := sm.client.Get(loginURL)
	if err != nil {
		return "", fmt.Errorf("fetch login page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read login page body: %w", err)
	}

	re := regexp.MustCompile(`name=['"]__csrf_magic['"]\s+value=['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", fmt.Errorf("CSRF token not found in login page")
	}
	token := matches[1]
	return token, nil
}

// isLoginValid 辅助方法：检查当前 client 是否能访问 graph_view.php（不修改状态）
func (sm *SessionManager) isLoginValid() bool {
	testURL := sm.baseURL + "/graph_view.php"
	resp, err := sm.client.Get(testURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	global.LOG.Debug(string(body))
	return !(bytes.Contains(body, []byte("登录到Cacti")) || bytes.Contains(body, []byte("<title>登录")))
}

// EnsureLogin 线程安全地确保已登录：若未登录则调用 ForceLogin()
// 推荐在每次业务请求前调用
func (sm *SessionManager) EnsureLogin() error {

	sm.loginMu.RLock()
	if sm.isLoggedIn {
		sm.loginMu.RUnlock()
		return nil
	}
	sm.loginMu.RUnlock()

	// 获取写锁并双检
	sm.loginMu.Lock()
	defer sm.loginMu.Unlock()
	if sm.isLoggedIn {
		return nil
	}

	if sm.isLoginValid() {
		global.LOG.Debug("logined")
		sm.isLoggedIn = true
		return nil
	}

	return sm.ForceLogin()
}

// Invalidate 主动使当前 session 失效（如密码变更、主动登出）
func (sm *SessionManager) Invalidate() {
	sm.loginMu.Lock()
	defer sm.loginMu.Unlock()
	sm.isLoggedIn = false
	global.LOG.Info("🔒 Session invalidated manually")
}

// GetLoginCount 返回累计成功登录次数（可用于监控）
func (sm *SessionManager) GetLoginCount() uint64 {
	sm.loginMu.RLock()
	defer sm.loginMu.RUnlock()
	return sm.loginCount
}
