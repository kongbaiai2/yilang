package routers

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/global"

	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"

	"github.com/gin-gonic/gin"
)

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		var body string
		if c.Request.Method == http.MethodPost {
			body = getRequestBody(c)
		}
		c.Next()

		cost := time.Since(start)
		var response interface{}
		value, exists := c.Get("cfw_code")
		if exists && c.Writer.Status() == http.StatusInternalServerError {
			response = value
		} else {
			response = c.Writer.Status()
		}
		global.LOG.Infof("[ACCESS] path:%s, response:%d, method:%s, query:%s, body:%s, ip:%s, user-agent:%s, cost:%dms",
			path, response, c.Request.Method, query, body, c.ClientIP(), c.Request.UserAgent(), cost.Milliseconds())
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					global.LOG.Errorf("url:%s, error:%s, request:%s", c.Request.URL.Path, err, string(httpRequest))
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					global.LOG.Errorf("[Recovery from panic] error:%s, request:%s, stack:%s", err, string(httpRequest), string(debug.Stack()))
				} else {
					global.LOG.Errorf("[Recovery from panic] error:%s request:%s", err, string(httpRequest))
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

type authHeader struct {
	RealIp      string `header:"X-Real-IP"`
	ContentType string `header:"Content-Type"`
}

type authParams struct {
	Identity string `form:"identity" binding:"required"`
	Bid      string `form:"bid" binding:"required"`
	Qtime    string `form:"qtime" binding:"required"`
	Sign     string `form:"sign" binding:"required"`
	RegionNo string `form:"regionNo"`
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := authHeader{}
		c.ShouldBindHeader(&h)

		remoteAddr := h.RealIp
		if remoteAddr == "" {
			remoteAddr = c.RemoteIP()
		}

		if strings.HasPrefix(remoteAddr, "127.0.0.1") {
			return
		}

		para := authParams{}

		authFailed := func() {
			global.LOG.Warnf("[CHECK] Authentication failed, param: %+v", para)
			e := errcode.ErrorAuthentication
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    e.Code,
				"success": false,
				"msg":     e.Error(),
				"data":    nil,
			})
		}

		var rawquery string
		if strings.EqualFold(c.Request.Method, "POST") &&
			strings.EqualFold(c.ContentType(), "application/x-www-form-urlencoded") {
			rawquery = getRequestBody(c)
		} else {
			rawquery = c.Request.URL.RawQuery
		}
		if rawquery == "" {
			global.LOG.Infof("[CHECK] APIAuth RawQuery is nil")
			authFailed()
			return
		}

		if err := c.ShouldBind(&para); err != nil {
			authFailed()
			return
		}

		// salt, bid, err := model.GetAuthInfo(global.DB, para.Identity)
		// if err != nil {
		// 	global.LOG.Infof("[CHECK] param error, no salt for identity: '%s', err: %s", para.Identity, err.Error())
		// 	authFailed()
		// 	return
		// }
		// if salt == "" {
		// 	global.LOG.Infof("[CHECK] param error, no default salt, no salt for identity: %s", para.Identity)
		// 	authFailed()
		// 	return
		// }

		// if !strings.EqualFold(bid, para.Bid) {
		// 	global.LOG.Infof("[CHECK] param error, bad bid:%s for identity:%s", para.Bid, para.Identity)
		// 	authFailed()
		// 	return
		// }

		qtimeInt, err := strconv.ParseInt(para.Qtime, 10, 64)
		if err != nil {
			global.LOG.Infof("[CHECK] APIAuth param error, bad qtime: %s", para.Qtime)
			authFailed()
			return
		}

		nowInt := time.Now().Unix()
		diff := nowInt - qtimeInt
		const MAX_AUTH_INTERVAL = 600
		if diff > MAX_AUTH_INTERVAL {
			global.LOG.Infof("[CHECK] APIAuth qtime diff: %v, max: %d, now:%d, qtime:%d", diff, MAX_AUTH_INTERVAL, nowInt, qtimeInt)
			authFailed()
			return
		}

		query, err := url.QueryUnescape(rawquery)
		if nil != err {
			global.LOG.Infof("[CHECK] APIAuth decode query failed")
			authFailed()
			return
		}

		pos := strings.LastIndex(query, "&sign=")
		if pos < 0 {
			global.LOG.Infof("[CHECK] APIAuth param error, sign required: %s", c.Request.URL.RawQuery)
			authFailed()
			return
		}

		// signStr := fmt.Sprintf("%s&salt=%s", query[:pos], salt)
		// signResult, err := utils.CalcSign(signStr)
		// if err != nil {
		// 	global.LOG.Infof("[CHECK] md5 calc failed, error: %s", err.Error())
		// 	authFailed()
		// 	return
		// }

		// if signResult != para.Sign {
		// 	global.LOG.Infof("[CHECK] APIAuth sign mismatch, expectSign: %s, paramSign: %s, calcMd5: %s", signResult, para.Sign, signStr)
		// 	authFailed()
		// 	return
		// }

		// if global.CONFIG.Region != para.RegionNo {
		// 	global.LOG.Infof("[INFO] API endpoint mismatch, endpointRegion: %s, requestRegion: %s", global.CONFIG.Region, para.RegionNo)
		// 	e := errcode.ErrorMethodNotSupport
		// 	if para.RegionNo == "" {
		// 		e.Msg = "The parameter regionNo is required."
		// 	} else {
		// 		e.Msg = "The specified endpoint can't operate this region."
		// 	}
		// 	c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
		// 		"code":    e.Code,
		// 		"success": false,
		// 		"msg":     e.Msg,
		// 		"data":    nil,
		// 	})
		// 	return
		// }
	}
}

func getRequestBody(c *gin.Context) string {
	data, err := c.GetRawData()
	if err != nil {
		global.LOG.Errorf("read body err:%s", err.Error())
	}

	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	return string(data)
}
