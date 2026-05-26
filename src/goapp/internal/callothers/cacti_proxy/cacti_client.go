package cacti_proxy

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"gonum.org/v1/gonum/stat"
)

var Cacti *CactiOptions

type CactiConfig struct {
	URL      string
	Username string
	Password string
}
type CactiOptions struct {
	cfg        CactiConfig
	client     *http.Client
	graph      *Graph
	sessionMgr *SessionManager
}
type Graph struct {
	localGraphID   int
	dataUrl        string
	start          int64
	end            int64
	filenamePrefix string
	isDownloaded   bool
	downUrl        string
}

func (g *Graph) Set(localGraphID int, start, end int64, filenamePrefix string, isDownloaded bool) *Graph {
	g.localGraphID = localGraphID
	g.start = start
	g.end = end
	g.filenamePrefix = filenamePrefix
	g.isDownloaded = isDownloaded
	return g
}

func (c_opt *CactiOptions) SetConfig(cfg CactiConfig) *CactiOptions {
	c_opt.client = createHTTPClient()
	c_opt.cfg = cfg
	c_opt.sessionMgr = NewSessionManager(cfg.URL, cfg.Username, cfg.Password, c_opt.client)

	return c_opt
}

func (c_opt *CactiOptions) FlushLogin()  {
	c_opt.sessionMgr.FlushLogin()
}
func (c_opt *CactiOptions) GetBaseURL() string {
	return c_opt.cfg.URL
}

func (c_opt *CactiOptions) LoginCacti() error {
	login_url := c_opt.cfg.URL + "/index.php"
	token, err := extractCSRFToken(c_opt.client, login_url)
	if err != nil {
		global.LOG.Errorf("Failed to extract CSRF token: %v", err)
		return err
	}
	return loginCacti(c_opt.client, login_url, c_opt.cfg.Username, c_opt.cfg.Password, token)
}

func (c_opt *CactiOptions) SetGraph(graph *Graph) *CactiOptions {
	c_opt.graph = graph
	return c_opt
}

func (g *Graph) FetchCustomData(client *http.Client, cut int) ([]float64, error) {
	body, err := g.FetchGraphData(client)
	if err != nil {
		return nil, err
	}
	return g.getData(body, cut)
}

// 创建带 Cookie 支持和跳过 TLS 验证的 HTTP 客户端
func createHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	jar, _ := cookiejar.New(nil)

	return &http.Client{
		Transport: tr,
		Jar:       jar,
		Timeout:   300 * time.Second,
	}
}

// 从登录页面提取 CSRF token (__csrf_magic)
func extractCSRFToken(client *http.Client, loginURL string) (string, error) {
	resp, err := client.Get(loginURL)
	if err != nil {
		global.LOG.Errorf("Failed to fetch login page: " + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`name=['"]__csrf_magic['"]\s+value=['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", fmt.Errorf("CSRF token not found in login page")
	}
	token := matches[1]
	// fmt.Println("✅ Got CSRF token:", token[:30]+"...")
	return token, nil
}

// 执行 Cacti 登录
func loginCacti(client *http.Client, loginURL, username, password, csrfToken string) error {
	loginData := url.Values{}
	loginData.Set("__csrf_magic", csrfToken)
	loginData.Set("action", "login")
	loginData.Set("login_username", username)
	loginData.Set("login_password", password)

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(loginData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", loginURL)

	loginResp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Login request failed: %s" + err.Error())
	}
	loginResp.Body.Close()

	// 验证是否登录成功
	testURL := strings.TrimSuffix(loginURL, "/index.php") + "/graph_view.php"
	testResp, err := client.Get(testURL)
	if err != nil {
		return fmt.Errorf("Test request after login failed: %s" + err.Error())
	}
	defer testResp.Body.Close()

	testBody, _ := io.ReadAll(testResp.Body)
	if bytes.Contains(testBody, []byte("登录到Cacti")) || bytes.Contains(testBody, []byte("<title>登录")) {
		return fmt.Errorf("❌ Login failed! Still on login page.")
	}
	// fmt.Println("✅ Login successful!")
	return nil
}

func (g *Graph) FetchGraphData(client *http.Client) ([]byte, error) {
	return fetchGraphData(client, g.dataUrl)
}

// 从 graph_xport.php 获取并解析 Inbound 数据（健壮版）
func fetchGraphData(client *http.Client, dataUrl string) ([]byte, error) {
	exportResp, err := client.Get(dataUrl)
	if err != nil {
		return nil, fmt.Errorf("Export request failed: " + err.Error())
	}
	defer exportResp.Body.Close()

	if exportResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Export request returned status: " + exportResp.Status)
	}

	return io.ReadAll(exportResp.Body)
}

func (g *Graph) getData(exportBody []byte, data_num int) ([]float64, error) {

	// 调试预览
	// fmt.Printf("Export response preview:\n%.*s\n", 500, string(exportBody))

	lines := strings.Split(string(exportBody), "\n")
	var csvLines []string
	inDataSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检测数据起始行：必须包含 "日期" 且是 CSV 格式（以引号开头）
		if !inDataSection {
			if strings.HasPrefix(line, `"日期`) || strings.Contains(line, `"日期"`) {
				inDataSection = true
				csvLines = append(csvLines, line)
			}
			continue
		}

		// 一旦进入数据区，后续非空行都视为数据
		csvLines = append(csvLines, line)
	}

	if len(csvLines) == 0 {
		return nil, fmt.Errorf("No data section found (missing '日期' header)")
	}

	// 将干净的数据行重新组合
	csvContent := strings.Join(csvLines, "\n")
	reader := csv.NewReader(strings.NewReader(csvContent))
	reader.LazyQuotes = true // 允许宽松引号（关键！）

	records, err := reader.ReadAll()
	if err != nil {
		// 如果还是失败，打印出问题的片段用于调试
		global.LOG.Errorf("Failed to parse cleaned CSV. First 5 lines:\n%s",
			strings.Join(csvLines[:min(5, len(csvLines))], "\n"))
		return nil, fmt.Errorf("Failed to parse CSV after cleaning: " + err.Error())
	}

	var inboundValues []float64
	for i, record := range records {
		if i == 0 {
			// 跳过标题行（"日期","Iutbound","Outbound"）
			continue
		}

		if len(record) < 3 {
			continue
		}

		if data_num > len(record)-1 {
			data_num = len(record) - 1
		}

		inValStr := strings.TrimSpace(record[data_num-1])
		if inValStr == "" {
			continue
		}

		if val, err := strconv.ParseFloat(inValStr, 64); err == nil && val > 0 {
			inboundValues = append(inboundValues, val)
		}
	}

	if len(inboundValues) == 0 {
		return nil, fmt.Errorf("No valid inbound values found in graph data")
	}

	// fmt.Printf("📈 Parsed %d valid inbound data points.\n", len(inboundValues))
	return inboundValues, nil
}

func (g *Graph) DownloadGraphImage(client *http.Client, filename string) error {
	// global.LOG.Errorf("graph: %+v", g)
	return downloadGraphImage(client, g.downUrl, filename)
}

// // 计算 95th 百分位
func calculateP95(values []float64) float64 {
	if len(values) == 0 {
		global.LOG.Errorf("empty values slice")
		return 0.0
	}
	sort.Float64s(values)
	// Empirical 分布 + 线性插值（默认行为，等价于 R-7 方法）
	return stat.Quantile(0.95, stat.Empirical, values, nil)
}

// 下载并保存图表图片
func downloadGraphImage(client *http.Client, graphImageURL, filename string) error {
	// fmt.Printf("📥 Downloading graph image from: %s\n", graphImageURL)

	resp, err := client.Get(graphImageURL)
	if err != nil {
		return fmt.Errorf("⚠️ Failed to download graph image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("⚠️ Graph image request failed with status: %s", resp.Status)
	}

	outFile, err := os.Create("img/" + filename)
	if err != nil {
		return fmt.Errorf("⚠️ Failed to create file %s: %v", filename, err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("⚠️ Failed to write image to file %s: %v", filename, err)
	}

	// fmt.Printf("✅ Graph image saved as: %s\n", filename)
	return nil
}

// 从 graph_xport.php 导出数据中提取 95th 百分位（远程计算值）
func extractCactiP95FromExport(body string) (float64, bool) {
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		if strings.Contains(line, "95th Percentile") || strings.Contains(line, "百分之百") {
			reader := csv.NewReader(strings.NewReader(line))
			reader.LazyQuotes = true
			record, err := reader.Read()
			if err != nil || len(record) < 2 {
				continue
			}
			if val, err := strconv.ParseFloat(record[1], 64); err == nil {
				return val, true
			}
		}
	}
	return 0, false
}

// 辅助函数：min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 主函数

// cactiURL := "https://.com:444/cacti"
// username := "admin"
// password := "pwd"
// localGraphID := 985

// ProcessMonthly(cactiURL, username, password, localGraphID)
// ProcessDaily(cactiURL, username, password, localGraphID)

func (c *CactiOptions) Do(g *Graph) (float64, error) {
	if err := c.sessionMgr.EnsureLogin(); err != nil {
		return 0, fmt.Errorf("ensure login failed: %w", err)
	}

	g.dataUrl = fmt.Sprintf("%s/graph_xport.php?local_graph_id=%d&rra_id=0&view_type=tree&graph_start=%d&graph_end=%d",
		c.cfg.URL, g.localGraphID, g.start, g.end)

	c.SetGraph(g)

	// 下载图表
	if g.isDownloaded {
		g.downUrl = fmt.Sprintf("%s/graph_image.php?local_graph_id=%d&rra_id=0&graph_start=%d&graph_end=%d",
			c.cfg.URL, g.localGraphID, g.start, g.end)
		filename := g.filenamePrefix + ".png"

		// global.LOG.Errorf("print c_opt: %+v", c)

		err := c.graph.DownloadGraphImage(c.client, filename)
		if err != nil {
			global.LOG.Errorf("Failed to fetch graph data: %v", err)
			// return 0, err
		}
	}

	// 获取并处理数据,获取95值
	allValues, err := c.graph.FetchGraphData(c.client)
	if err != nil {
		// ⚠️ 可选：session 可能刚过期？再检一次并重试
		if !c.sessionMgr.IsAlive() {
			c.sessionMgr.Invalidate()
			if retryErr := c.sessionMgr.EnsureLogin(); retryErr == nil {
				allValues, err = g.FetchGraphData(c.client)
			}
		}
		if err != nil {
			return 0, fmt.Errorf("fetch data failed: %w", err)
		}
	}
	// // inbound=2 oubound=3
	// inboundValues, err := c.graph.getData(allValues, 2)
	// if err != nil {
	// 	global.LOG.Errorf("Failed to fetch graph data: %v", err)
	// 	global.LOG.Errorf("c.graph: %+v", c.graph)
	// 	return 0, err
	// }

	// p95 := calculateP95(inboundValues)

	remote_p95, _ := extractCactiP95FromExport(string(allValues))

	// if (p95-remote_p95) > 1000 || (remote_p95-p95) > 1000 {
	// 	global.LOG.Errorf("Remote report 95th place - calculation of 95th percentile data with significant difference.remote_p95:%f, p95:%f ", remote_p95, p95)
	// }

	return remote_p95, nil
}

// FetchOption 可选参数
// base := "https://api.example.com/data"

// final := ApplyOptions(base,
//     WithQueryParam("format", "json"),
//     WithQueryParam("limit", "100"),
//     WithHeader("Authorization", "Bearer xyz"),
// )

// fmt.Println("Final URL:", final)
type FetchOption func(string) string

func WithStart(start string) FetchOption {
	return func(u string) string {
		return u + "&start=" + start
	}
}

func WithEnd(end string) FetchOption {
	return func(u string) string {
		return u + "&end=" + end
	}
}

func WithStep(step int) FetchOption {
	return func(u string) string {
		return u + "&step=" + strconv.Itoa(step)
	}
}

// ✅ Step 2: 创建具体选项函数
func WithQueryParam(key, value string) FetchOption {
	return func(u string) string {
		parsed, err := url.Parse(u)
		if err != nil {
			return u
		}
		q := parsed.Query()
		q.Set(key, value)
		parsed.RawQuery = q.Encode()
		return parsed.String()
	}
}

func WithHeader(name, value string) FetchOption {
	// 演示：虽然不改 URL，但可扩展为传 context 或 options struct
	// 这里我们只打印示意（实际中可能封装进 client）
	// fmt.Printf("[INFO] Will set header %s: %s\n", name, value)
	return func(u string) string { return u } // 不修改 URL
}

// ✅ Step 3: 应用所有选项
func ApplyOptions(base string, opts ...FetchOption) string {
	for _, opt := range opts {
		base = opt(base)
	}
	return base
}
