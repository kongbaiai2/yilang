package routers

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"log"
// 	"net"
// 	"net/http"
// 	"net/http/httptest"
// 	"reflect"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"

// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/agiledragon/gomonkey"
// 	sls "github.com/aliyun/aliyun-log-go-sdk"
// 	"github.com/aliyun/aliyun-oss-go-sdk/oss"
// 	"github.com/gin-gonic/gin"
// 	. "github.com/smartystreets/goconvey/convey"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"
// )

// func initLog() {
// 	zapConf := middleware.ZapConf{
// 		Level:       global.CONFIG.Zap.Level,
// 		Path:        global.CONFIG.Zap.Path,
// 		Format:      global.CONFIG.Zap.Format,
// 		Prefix:      global.CONFIG.Zap.Prefix,
// 		EncodeLevel: global.CONFIG.Zap.EncodeLevel,
// 	}

// 	logrotateConf := middleware.LogrotateConf{
// 		MaxSize:    global.CONFIG.LogRotate.MaxSize,
// 		MaxBackups: global.CONFIG.LogRotate.MaxBackups,
// 		MaxAges:    global.CONFIG.LogRotate.MaxAges,
// 		Compress:   global.CONFIG.LogRotate.Compress,
// 	}

// 	global.LOG = middleware.Zap(zapConf, logrotateConf)
// }

// func setupRouter() *gin.Engine {
// 	v := middleware.ViperParseConf("../../../config.yaml")
// 	if err := v.Unmarshal(&global.CONFIG); err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	initLog()

// 	gin.SetMode(gin.TestMode)
// 	r := NewRouter()
// 	return r
// }

// type HttpResp struct {
// 	Msg string `json:"message"`
// }

// func TestHealthCheck(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest("GET", "/status.default", nil)
// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, 200, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, "pong", resp.Msg)
// }

// func TestHealthCheck_ServiceUnavailable(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest("GET", "/status.default", nil)
// 	global.ProcessExit = true
// 	r := setupRouter()
// 	r.ServeHTTP(w, req)
// 	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, "service unavailable", resp.Msg)
// }

// func TestCreateRsPool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	//body
// 	body := "regionNo=cn-huhehaote-test-3&rsPoolName=test&rsPoolBiz=cfw&resGroupName=&flowLogLogstore=40c16f05-5a53-4349-a4a4-caf8fa7da1f2&ipsRulesBucket=11&ipGeoBucket=11&userConfigBucket=22&bypassConfig=%7B%7D&dryRun=true&identity=secnet&bid=secnet&qtime=1679393919&sign=fdbbc91f235198a5a79fcb23c3abb222"
// 	req := httptest.NewRequest("POST", "/ops/rspool/create", strings.NewReader(body))

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

// 	patches := gomonkey.ApplyFunc(
// 		model.AddRspool,
// 		func(*gorm.DB, *model.RsPool, *model.RsPoolBypass) error {
// 			return nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.GetLogstores,
// 		func(*gorm.DB, []string) (*[]model.Logstore, error) {
// 			return &[]model.Logstore{{LogType: resource.LogTypeCfwFlowLog}}, nil
// 		})

// 	patches.ApplyFunc(
// 		segw.CreateRsPool,
// 		func(string, string) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestCreatePrivateDnsEndpoint(t *testing.T) {

// 	r := setupRouter() // 初始化路由
// 	global.DB = model.InitDB(global.CONFIG.Mysql)

// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest("GET", "/api/privateDns/CreatePrivateDnsEndpoint?regionNo=cn-hangzhou&aliUid=1358090477156271&primaryVSwitchId=vsw-2ze0krvvikv3v4sj375t5&firewallType=[\"vpc\"]&accessInstanceName=xxx&privateDnsType=Custom&vpcId=vpc-2zec9ztnlz6s1fyt3jhw1&primaryVSwitchIp=192.168.1.100&port=53&ipProtocol=UDP&primaryDns=100.100.2.138", nil)

// 	// 设置请求头
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.AddPrivateDns,
// 		func(*gorm.DB, *model.PrivateDns) error {
// 			return nil
// 		},
// 	)

// 	patches.ApplyFunc(
// 		vpc_proxy.DescribeVSwitchesApiProxy,
// 		func(uint64, string, string, string) ([]vpc.VSwitch, *errcode.Err) {
// 			return []vpc.VSwitch{{
// 				VpcId:                   "vpc-2zec9ztnlz6s1fyt3jhw1",
// 				Status:                  "Available",
// 				CreationTime:            "2023-01-01T00:00:00Z",
// 				IsDefault:               true,
// 				AvailableIpAddressCount: 10,
// 				NetworkAclId:            "acl-12345678",
// 				OwnerId:                 123456789,
// 				VSwitchId:               "vsw-2ze0krvvikv3v4sj375t5",
// 				CidrBlock:               "192.168.1.0/24",
// 				Description:             "Sample VSwitch 1",
// 				ResourceGroupId:         "rg-12345678",
// 				ZoneId:                  "us-west-1a",
// 				Ipv6CidrBlock:           "2001:db8::/32",
// 				VSwitchName:             "MyVSwitch1",
// 				ShareType:               "NONE",
// 				EnabledIpv6:             true,
// 				RouteTable:              vpc.RouteTable{},              // 如果需要，添加实际的 RouteTable
// 				Tags:                    vpc.TagsInDescribeVSwitches{}, // 如果需要添加标签，确保这个结构体也是实例化的
// 			}}, errcode.StatusSuccess
// 		},
// 	)

// 	patches.ApplyFunc(
// 		model.GetAvailableNatIpServiceByZoneIdStatus,
// 		func(*gorm.DB, string, string) ([]model.NatGatewayIpService, error) {
// 			return []model.NatGatewayIpService{{
// 				RegionNo:          "cn-hangzhou",
// 				NatGatewayId:      "nat-123456",
// 				VpcId:             "vpc-2zec9ztnlz6s1fyt3jhw1",
// 				ZoneId:            "us-west-1a",
// 				VSwitchId:         "vswitch-123456",
// 				NatIpCidr:         "192.168.1.0/24",
// 				NatIp:             "192.168.1.1",
// 				Status:            "Available",
// 				Description:       "Main NAT gateway IP service",
// 				EndpointServiceId: "es-123456",
// 			}}, nil
// 		},
// 	)

// 	patches.ApplyFunc(
// 		model.GetPrivateDnsCountByFirewallType,
// 		func(*gorm.DB, uint64, string, string) (int64, error) {
// 			return 0, nil
// 		},
// 	)

// 	patches.ApplyFunc(
// 		model.AddPrivateDnsTask,
// 		func(*gorm.DB, *model.PrivateDnsTask) error {
// 			return nil
// 		},
// 	)

// 	patches.ApplyFunc(
// 		model.GetPrivateDnsCountByAccessId,
// 		func(*gorm.DB, string) (int64, error) {
// 			return 0, nil
// 		},
// 	)

// 	patches.ApplyFunc(
// 		task.InitTaskProgressInfo,
// 		func(*gorm.DB, uint64, string, string) (*task.TaskProgressInfo, *errcode.Err) {
// 			return &task.TaskProgressInfo{}, errcode.StatusSuccess
// 		},
// 	)

// 	patches.ApplyFunc(
// 		model.AddAccessInstanceTaskProgress,
// 		func(*gorm.DB, *model.AccessInstanceTaskProgress) error {
// 			return nil
// 		},
// 	)

// 	defer patches.Reset()

// 	r.ServeHTTP(w, req) // 处理请求

// 	// 检查响应状态码
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	// 解析返回的 JSON 响应
// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.NoError(t, err)                               // 检查解析错误
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg) // 验证消息
// }

// func TestDeleteRsPool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/rspool/delete?regionNo=cn-hangzhou&rsPoolId=xxx-xxx-xxx", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.GetRspoolByID,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{}, nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		segw.DeleteRsPool,
// 		func(string, string) error {
// 			return nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetUserRspoolByRegionAndPoolName,
// 		func(*gorm.DB, *model.RedirectUserRspool) (*model.RedirectUserRspool, error) {
// 			return nil, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetRedirectRspoolByRegionAndName,
// 		func(*gorm.DB, *model.RedirectRspool) (*model.RedirectRspool, error) {
// 			return nil, nil
// 		})

// 	patches.ApplyFunc(
// 		model.DeleteRspool,
// 		func(*gorm.DB, string) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestUpdateRsPool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	//body
// 	body := "regionNo=cn-huhehaote-test-3&rsPoolBiz=cfw&rsPoolId=xxx-xxx-xxx&flowLogLogstore=12&trafficLogLogstore=33&ipsHitLogstore=22&aclHitLogstore=22&ipsRulesBucket=11&ipGeoBucket=11&userConfigBucket=22"
// 	req := httptest.NewRequest("POST", "/ops/rspool/update", strings.NewReader(body))

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

// 	patches := gomonkey.ApplyFunc(
// 		model.UpdateRspool,
// 		func(*gorm.DB, *model.RsPool) error {
// 			return nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.GetRspoolByID,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{RsPoolBiz: rspool.RsPoolBizTypeCfw}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetLogstores,
// 		func(*gorm.DB, []string) (*[]model.Logstore, error) {
// 			return &[]model.Logstore{{}}, nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusInternalServerError, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.ErrorParameters.Msg, resp.Msg)
// }

// func TestUpdateRsPoolBypass(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	//body
// 	body := "regionNo=cn-huhehaote-test-3&rsPoolId=xxx-xxx-xxx&globalConfig={}&userConfig={}&eipConfig={}"
// 	req := httptest.NewRequest("POST", "/ops/rspool/bypass/update", strings.NewReader(body))

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

// 	patches := gomonkey.ApplyFunc(
// 		model.UpdateRsPoolBypassByRsPoolId,
// 		func(*gorm.DB, interface{}, string) error {
// 			return nil
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestUpdateRsPoolAttribute(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/rspool/attribute/update?regionNo=cn-huhehaote-test-3&rsPoolId=xxx-xxx-xxx&tags=[\"a\",\"b\"]&config={\"a\":\"a1\"}", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

// 	patches := gomonkey.ApplyFunc(
// 		model.GetRspoolByID,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{Attr: "{}"}, nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.UpdateRspoolAttribute,
// 		func(*gorm.DB, *model.RsPool, string) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDescribeRsPool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/rspool/describe?regionNo=cn-hangzhou", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		rspool.GetRspoolList,
// 		func(*rspool.DescribeRspoolRequest) (cnt int64, rsPools []model.RsPool, err error) {
// 			return
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.GetAllLogstores,
// 		func(*gorm.DB) (*[]model.Logstore, error) {
// 			return &[]model.Logstore{}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetAllBuckets,
// 		func(*gorm.DB) (*[]model.Bucket, error) {
// 			return &[]model.Bucket{}, nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// //func TestAddRs(t *testing.T) {
// //	w := httptest.NewRecorder()
// //
// //	//body
// //	body := "regionNo=cn-huhehaote-test-3&rsPoolId=xxx-xxx&rsIpList=[{\"ip_addr\":\"11.0.0.100\",\"weight\":1},{\"ip_addr\":\"12.0.0.100\",\"weight\":1},{\"ip_addr\":\"13.0.0.100\",\"weight\":1}]&healthCheck=10"
// //	req := httptest.NewRequest("POST", "/ops/rs/add", strings.NewReader(body))
// //
// //	//header
// //	req.Header.Add("X-Real-IP", "127.0.0.1")
// //	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// //
// //	patches := gomonkey.ApplyFunc(
// //		model.GetRspoolByID,
// //		func(*gorm.DB, string) (*model.RsPool, error) {
// //			return &model.RsPool{}, nil
// //		})
// //	defer patches.Reset()
// //
// //	patches.ApplyFunc(
// //		segw.AddRs,
// //		func(string, string, string) error {
// //			return nil
// //		})
// //
// //	patches.ApplyFunc(
// //		model.AddRsList,
// //		func(*gorm.DB, *[]model.Rs) error {
// //			return nil
// //		})
// //
// //	r := setupRouter()
// //	r.ServeHTTP(w, req)
// //
// //	assert.Equal(t, http.StatusOK, w.Code)
// //
// //	resp := HttpResp{}
// //	err := json.Unmarshal(w.Body.Bytes(), &resp)
// //	assert.ErrorIs(t, err, nil)
// //	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// //}

// func TestUpdateRs(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	//body
// 	body := "regionNo=cn-huhehaote-test-3&rsPoolId=xxx-xxx&rsIpList=1.1.1.1,2.2.2.2&healthCheck=20"
// 	req := httptest.NewRequest("POST", "/ops/rs/update", strings.NewReader(body))

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")
// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

// 	patches := gomonkey.ApplyFunc(
// 		model.UpdateRsByIpList,
// 		func(*gorm.DB, map[string]interface{}, string, []string) error {
// 			return nil
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDeleteRs(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/rs/delete?regionNo=cn-shanghai&rsPoolId=9a548630-f902-45ce-9127-1845855b5c25&rsIpList=[{\"ip_addr\":\"11.0.0.100\",\"weight\":1},{\"ip_addr\":\"12.0.0.100\",\"weight\":1},{\"ip_addr\":\"13.0.0.100\",\"weight\":1}]", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.GetRspoolByID,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{DryRun: false, RsPoolBiz: "cfw"}, nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		segw.DeleteRs,
// 		func(string, string, string) error {
// 			return nil
// 		})

// 	patches.ApplyFunc(
// 		model.DeleteRsByIpList,
// 		func(*gorm.DB, string, []string) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDescribeRs(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/rs/describe?regionNo=cn-shanghai&rsPoolId=9a548630-f902-45ce-9127-1845855b5c25", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		realserver.GetRsList,
// 		func(*realserver.DescribeRsInnerRequest) (cnt int64, rslist []model.Rs, err error) {
// 			return
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDescribeRsHost(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest("GET", "/ops/rs/host/describe?ip=11.207.249.82", nil)
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		skyline.GetServerInfo,
// 		func(opt *skyline.GetServerInfoRequest) (*skyline.GetServerInfoResponse, error) {
// 			return &skyline.GetServerInfoResponse{}, nil
// 		})

// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	// 检查响应状态码
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	// 解析返回的 JSON 响应
// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestCreateBucket(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/resource/createBucket?regionNo=cn-shanghai&bucket=testBucket", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		oss.New,
// 		func(endpoint, accessKeyID, accessKeySecret string, options ...oss.ClientOption) (*oss.Client, error) {
// 			return &oss.Client{}, nil
// 		})
// 	defer patches.Reset()

// 	var client oss.Client
// 	patches.ApplyMethod(
// 		reflect.TypeOf(client),
// 		"CreateBucket",
// 		func(oss.Client, string, ...oss.Option) error {
// 			return nil
// 		})
// 	patches.ApplyFunc(
// 		model.AddBucket,
// 		func(*gorm.DB, *model.Bucket) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestCreateLogstore(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/resource/createLogstore?regionNo=cn-shanghai&project=testProject&logstore=testLogStore&logType=flow_log", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		sls.CreateNormalInterface,
// 		func(string, string, string, string) sls.ClientInterface {
// 			return &sls.Client{}
// 		})
// 	defer patches.Reset()

// 	var client *sls.Client
// 	patches.ApplyMethod(
// 		reflect.TypeOf(client),
// 		"CreateProject",
// 		func(*sls.Client, string, string) (project *sls.LogProject, err error) {
// 			return
// 		})
// 	patches.ApplyMethod(
// 		reflect.TypeOf(client),
// 		"CreateLogStoreV2",
// 		func(*sls.Client, string, *sls.LogStore) (err error) {
// 			return
// 		})
// 	patches.ApplyFunc(
// 		model.AddLogstore,
// 		func(*gorm.DB, *model.Logstore) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDescribeLogStoreBucket(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/resource/describe?regionNo=cn-shanghai", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.GetAllLogstores,
// 		func(*gorm.DB) (*[]model.Logstore, error) {
// 			return &[]model.Logstore{}, nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.GetAllBuckets,
// 		func(*gorm.DB) (*[]model.Bucket, error) {
// 			return &[]model.Bucket{}, nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestAddDefaultRspool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/default_rspool/add?regionNo=cn-shanghai&rsPoolName=xxx&userType=buy&version=2", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.GetRspoolByName,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{}, nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.GetRedirectRspool,
// 		func(*gorm.DB, *model.RedirectRspool) (*model.RedirectRspool, error) {
// 			return nil, nil
// 		})
// 	patches.ApplyFunc(
// 		model.AddRedirectRspool,
// 		func(*gorm.DB, *model.RedirectRspool) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestSwitchDefaultRspool(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	req := httptest.NewRequest("GET", "/ops/redirect/default_rspool/switch?regionNo=cn-shanghai&rsPoolName=x1&oldRsPoolName=x2", nil)

// 	// 设置请求头
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	// 模拟两个函数的返回
// 	patches := gomonkey.ApplyFunc(
// 		model.GetRspoolByName,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{}, nil
// 		},
// 	)
// 	patches.ApplyFunc(
// 		model.UpdateUserInfoRsPool,
// 		func(*gorm.DB, string, *model.UserInfo) error {
// 			return nil
// 		},
// 	)
// 	defer patches.Reset()

// 	r := setupRouter()  // 初始化路由
// 	r.ServeHTTP(w, req) // 处理请求

// 	// 检查响应状态码
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	// 解析返回的 JSON 响应
// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.NoError(t, err)                               // 检查解析错误
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg) // 验证消息
// }

// func TestDescribeDefaultRspool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/default_rspool/describe?regionNo=cn-shanghai&rsPoolName=xxx&userType=buy&version=2", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		redirect.GetRedirectRspoolList,
// 		func(*redirect.DescribeDefaultRspoolRequest) (c int64, p []model.RedirectRspool, e error) {
// 			return
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDeleteDefaultRspool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/default_rspool/delete?regionNo=cn-shanghai&rsPoolName=xxx&userType=buy&version=2", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.DeleteRedirectRspool,
// 		func(*gorm.DB, *model.RedirectRspool) error {
// 			return nil
// 		})
// 	patches.ApplyFunc(
// 		model.GetRspoolByName,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return nil, nil
// 		})

// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestAddUserRsPool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/user_rspool/add?regionNo=cn-shanghai&rsPoolName=xxx&aliUid=123&ipVersion=ipv4&eip=1.2.3.4", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.GetRspoolByName,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{}, nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.GetRedirectUserRspool,
// 		func(*gorm.DB, *model.RedirectUserRspool) (*model.RedirectUserRspool, error) {
// 			return nil, nil
// 		})
// 	patches.ApplyFunc(
// 		model.AddRedirectUserRspool,
// 		func(*gorm.DB, *model.RedirectUserRspool) error {
// 			return nil
// 		})

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDeleteUserRsPool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/user_rspool/delete?regionNo=cn-shanghai&rsPoolName=xxx&aliUid=123&ipVersion=ipv4&eip=1.2.3.4", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		model.DeleteRedirectUserRspool,
// 		func(*gorm.DB, *model.RedirectUserRspool) error {
// 			return nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetRspoolByName,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return nil, nil
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestDescribeUserRsPool(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/user_rspool/describe?regionNo=cn-shanghai", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		redirect.GetRedirectUserRspoolList,
// 		func(*redirect.DescribeUserRspoolRequest) (c int64, p []model.RedirectUserRspool, e error) {
// 			return
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestOpsAddRule(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/rule/add?regionNo=cn-shanghai&rsPoolName=xxx&aliUid=123&ipVersion=ipv4&eip=1.2.3.4", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		segw.AddRedirectRuleSync,
// 		func(a, b, c, d, e string) error {
// 			return nil
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestOpsDeleteRule(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/rule/delete?regionNo=cn-shanghai&rsPoolName=xxx&aliUid=123&ipVersion=ipv4&eip=1.2.3.4", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		segw.DeleteRedirectRuleSync,
// 		func(a, b, c string) error {
// 			return nil
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestOpsUpdateRule(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/rule/update?regionNo=cn-shanghai&rsPoolName=xxx&aliUid=123&ipVersion=ipv4&eip=1.2.3.4", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		segw.SwitchRspool,
// 		func(a, b, c, d string) error {
// 			return nil
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestOpsDescribeRule(t *testing.T) {
// 	w := httptest.NewRecorder()

// 	req := httptest.NewRequest("GET", "/ops/redirect/rule/describe?regionNo=cn-shanghai&rsPoolName=xxx&aliUid=123&ipVersion=ipv4&eip=1.2.3.4", nil)

// 	//header
// 	req.Header.Add("X-Real-IP", "127.0.0.1")

// 	patches := gomonkey.ApplyFunc(
// 		segw.QueryRedirectRule,
// 		func(a, b, c string) (segw.RedirectRuleItem, error) {
// 			return segw.RedirectRuleItem{}, nil
// 		})
// 	defer patches.Reset()

// 	r := setupRouter()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)

// 	resp := HttpResp{}
// 	err := json.Unmarshal(w.Body.Bytes(), &resp)
// 	assert.ErrorIs(t, err, nil)
// 	assert.Equal(t, errcode.StatusSuccess.Msg, resp.Msg)
// }

// func TestUserEnableFirewall(t *testing.T) {
// 	Convey("TestUserEnableFirewall", t, func() {
// 		Convey("Test Firewall Already Enabled", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/firewall/enable?regionNo=cn-shanghai&aliUid=123&ipVersion=ipv4&userType=trial&version=0&shareCluster=true", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			rows := sqlmock.NewRows([]string{"ShareDocker", "RsPool"}).
// 				AddRow(1, "cn-shanghai-test")
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").WillReturnRows(rows)
// 			mock.ExpectCommit()

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(http.StatusOK, ShouldEqual, w.Code)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(errcode.StatusSuccess.Msg, ShouldEqual, resp.Msg)
// 		})

// 		Convey("Test Normal Case", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/firewall/enable?regionNo=cn-shanghai&aliUid=123&ipVersion=ipv4&userType=trial&version=0&shareCluster=true", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			rows1 := sqlmock.NewRows([]string{"ShareDocker", "RsPool"})
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").WillReturnRows(rows1)
// 			rows2 := sqlmock.NewRows([]string{"RsPool"}).AddRow("cn-shanghai-trial")
// 			mock.ExpectQuery("select rs_pool from fw_rs_pool_dispatcher ").WillReturnRows(rows2)
// 			mock.ExpectExec("insert").WillReturnResult(sqlmock.NewResult(1, 1))
// 			mock.ExpectCommit()

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(w.Code, ShouldEqual, http.StatusOK)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 		})
// 	})
// }

// func TestUserDisableFirewall(t *testing.T) {
// 	Convey("TestUserDisableFirewall", t, func() {
// 		Convey("Test Firewall Already Disabled", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/firewall/disable?regionNo=cn-shanghai&aliUid=123", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			rows := sqlmock.NewRows([]string{"ShareDocker", "RsPool"})
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").WillReturnRows(rows)
// 			mock.ExpectCommit()

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(http.StatusOK, ShouldEqual, w.Code)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(errcode.StatusSuccess.Msg, ShouldEqual, resp.Msg)
// 		})

// 		Convey("Test Normal Case", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/firewall/disable?regionNo=cn-shanghai&aliUid=123", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			rows1 := sqlmock.NewRows([]string{"ShareDocker", "RsPool"}).
// 				AddRow(1, "RsPool")
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").
// 				WillReturnRows(rows1)

// 			rows2 := sqlmock.NewRows([]string{"cnt"})
// 			mock.ExpectQuery("select count").WillReturnRows(rows2)

// 			mock.ExpectExec("delete from fw_user_info").WillReturnResult(sqlmock.NewResult(1, 1))

// 			mock.ExpectCommit()

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(http.StatusOK, ShouldEqual, w.Code)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(errcode.StatusSuccess.Msg, ShouldEqual, resp.Msg)
// 		})
// 	})
// }

// func TestAddEIPRule(t *testing.T) {
// 	Convey("TestAddEIPRule", t, func() {
// 		Convey("Test Firewall Not Enabled", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/rule/add?regionNo=cn-shanghai&aliUid=123&ipVersion=ipv4&eip=[\"1.2.3.4\"]", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			fwNotEnabled := sqlmock.NewRows([]string{"ShareDocker", "RsPool"})
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").WillReturnRows(fwNotEnabled)

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(w.Code, ShouldEqual, http.StatusInternalServerError)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(resp.Msg, ShouldEqual, errcode.ErrorFirewallNotEnabled.Msg)
// 		})

// 		Convey("Test eip rule already added, and aliuid correct", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/rule/add?regionNo=cn-shanghai&aliUid=123&ipVersion=ipv4&eip=[\"1.2.3.4\"]", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			fwEnabled := sqlmock.NewRows([]string{"ShareDocker", "RsPool"}).
// 				AddRow("0", "cn-shanghai-test")
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").WillReturnRows(fwEnabled)

// 			UserRspoolNoRows := sqlmock.NewRows([]string{"IpVersion", "Eip", "RsPool"})
// 			mock.ExpectQuery("select ip_version, eip, rs_pool from fw_user_rs_pool").WillReturnRows(UserRspoolNoRows)

// 			EipRuleStatus2 := sqlmock.NewRows([]string{"Status"}).AddRow(2)
// 			mock.ExpectQuery("SELECT CASE WHEN").WillReturnRows(EipRuleStatus2)

// 			mock.ExpectCommit()

// 			mock.ExpectExec("replace into user_config_update_log").WillReturnResult(sqlmock.NewResult(1, 1))

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(w.Code, ShouldEqual, http.StatusOK)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(errcode.StatusSuccess.Msg, ShouldEqual, resp.Msg)
// 		})
// 	})
// }

// func TestListEIPRules(t *testing.T) {
// 	Convey("TestListEIPRules", t, func() {
// 		w := httptest.NewRecorder()

// 		req := httptest.NewRequest("GET", "/api/redirect/rule/describe?regionNo=cn-shanghai&aliUid=123&ipVersion=ipv4&eip=1.2.3.4", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		db, mock, err := sqlmock.New()
// 		if err != nil {
// 			t.Fatalf("failed to open sqlmock database: %s", err)
// 		}
// 		defer db.Close()

// 		global.SqlDB = db
// 		rows := sqlmock.NewRows([]string{"aliUid", "CreateTime", "Region", "Eip", "RsPool", "IpVersion"}).
// 			AddRow("123", 123, "cn-shanghai", "1.2.3.4", "rs_pool", "ipv4")
// 		mock.ExpectQuery("select").WillReturnRows(rows)

// 		global.DB, err = gorm.Open(mysql.New(mysql.Config{
// 			SkipInitializeWithVersion: true,
// 			Conn:                      db,
// 		}), &gorm.Config{})
// 		OneboxRedirectRule := sqlmock.NewRows([]string{"id", "aliuid", "regionId", "eip", "rsPoolName", "ipVersion"})
// 		mock.ExpectQuery("SELECT").WillReturnRows(OneboxRedirectRule)

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)

// 		resp := HttpResp{}
// 		err = json.Unmarshal(w.Body.Bytes(), &resp)
// 		So(err, ShouldEqual, nil)
// 		So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 	})
// }

// func TestDeleteEIPRule(t *testing.T) {
// 	Convey("TestDeleteEIPRule", t, func() {
// 		Convey("Test Firewall Not Enabled", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/rule/delete?regionNo=cn-shanghai&aliUid=123&ipVersion=ipv4&eip=[\"1.2.3.4\"]", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			fwNotEnabled := sqlmock.NewRows([]string{"ShareDocker", "RsPool"})
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").WillReturnRows(fwNotEnabled)

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(http.StatusOK, ShouldEqual, w.Code)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 		})

// 		Convey("Test no eip rule in db", func() {
// 			w := httptest.NewRecorder()

// 			req := httptest.NewRequest("GET", "/api/redirect/rule/delete?regionNo=cn-shanghai&aliUid=123&ipVersion=ipv4&eip=[\"1.2.3.4\"]", nil)

// 			//header
// 			req.Header.Add("X-Real-IP", "127.0.0.1")

// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatalf("failed to open sqlmock database: %s", err)
// 			}
// 			defer db.Close()

// 			global.SqlDB = db
// 			mock.ExpectBegin()
// 			fwEnabled := sqlmock.NewRows([]string{"ShareDocker", "RsPool"}).
// 				AddRow("0", "cn-shanghai-test")
// 			mock.ExpectQuery("select share_docker,rs_pool from fw_user_info").WillReturnRows(fwEnabled)

// 			noUserRule := sqlmock.NewRows([]string{"cnt"}).AddRow(0)
// 			mock.ExpectQuery("select").WillReturnRows(noUserRule)

// 			mock.ExpectCommit()

// 			mock.ExpectExec("replace into user_config_update_log").WillReturnResult(sqlmock.NewResult(1, 1))

// 			r := setupRouter()
// 			r.ServeHTTP(w, req)

// 			So(http.StatusOK, ShouldEqual, w.Code)

// 			resp := HttpResp{}
// 			err = json.Unmarshal(w.Body.Bytes(), &resp)
// 			So(err, ShouldEqual, nil)
// 			So(errcode.StatusSuccess.Msg, ShouldEqual, resp.Msg)
// 		})
// 	})
// }

// func TestRsPoolUserDescribe(t *testing.T) {
// 	Convey("TestRsPoolUserDescribe", t, func() {
// 		w := httptest.NewRecorder()

// 		req := httptest.NewRequest("GET", "/ops/emerg/user/describe?regionNo=cn-shanghai&rsPoolName=test&firewallType=InternetFirewall", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		db, mock, err := sqlmock.New()
// 		if err != nil {
// 			t.Fatalf("failed to open sqlmock database: %s", err)
// 		}
// 		defer db.Close()

// 		global.DB, err = gorm.Open(mysql.New(mysql.Config{
// 			SkipInitializeWithVersion: true,
// 			Conn:                      db,
// 		}), &gorm.Config{})
// 		So(err, ShouldEqual, nil)
// 		patches := gomonkey.ApplyFunc(
// 			model.GetRspoolByName,
// 			func(*gorm.DB, string) (*model.RsPool, error) {
// 				return &model.RsPool{}, nil
// 			})
// 		defer patches.Reset()

// 		rows := sqlmock.NewRows([]string{"cnt"}).AddRow(1)
// 		mock.ExpectQuery("SELECT").WillReturnRows(rows)
// 		rows = sqlmock.NewRows([]string{"region", "rsPool", "aliUid"}).AddRow("", "", "123")
// 		mock.ExpectQuery("SELECT").WillReturnRows(rows)

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)

// 		resp := HttpResp{}
// 		err = json.Unmarshal(w.Body.Bytes(), &resp)
// 		So(err, ShouldEqual, nil)
// 		So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 	})
// }

// func TestFirewallApi(t *testing.T) {
// 	patches := gomonkey.ApplyFunc(
// 		model.GetRsByBizIp,
// 		func(*gorm.DB, string, string) (*model.Rs, error) {
// 			return &model.Rs{Attr: "{\"tags\":[\"IPS_GRAY\"]}"}, nil
// 		})
// 	defer patches.Reset()

// 	patches.ApplyFunc(
// 		model.GetRspoolByID,
// 		func(*gorm.DB, string) (*model.RsPool, error) {
// 			return &model.RsPool{Attr: "{\"tags\":[\"IPv6Geo_GRAY\",\"IPGeo_GRAY\"],\"config\":{\"a\":\"a1\",\"b\":\"b1\"}}"}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetBypassConfigByRsPool,
// 		func(*gorm.DB, string) (model.RsPoolBypass, error) {
// 			return model.RsPoolBypass{GlobalAcl: true}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetFullConfigObject,
// 		func(*gorm.DB, string, string, []string) ([]model.FullObject, error) {
// 			return []model.FullObject{{
// 				RegionNo:   "cn-hangzhou",
// 				RsPoolName: "rs_pool",
// 				ConfigName: "firewall_config_name",
// 				BucketId:   "oss_config_id",
// 				Object:     "oss_path",
// 			}}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetBuckets,
// 		func(*gorm.DB, []string) (*[]model.Bucket, error) {
// 			return &[]model.Bucket{{
// 				RegionNo:        "cn-hangzhou",
// 				BucketId:        "oss_config_id",
// 				Endpoint:        "endpoint",
// 				Bucket:          "bucket",
// 				AccessKeyId:     "xxx",
// 				AccessKeySecret: "xxx",
// 			}}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetBucketById,
// 		func(*gorm.DB, string) (*model.Bucket, error) {
// 			return &model.Bucket{
// 				RegionNo:        "cn-hangzhou",
// 				BucketId:        "oss_config_id",
// 				Endpoint:        "endpoint",
// 				Bucket:          "bucket",
// 				AccessKeyId:     "xxx",
// 				AccessKeySecret: "xxx",
// 			}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetLogstores,
// 		func(*gorm.DB, []string) (*[]model.Logstore, error) {
// 			return &[]model.Logstore{{
// 				RegionNo:        "cn-hangzhou",
// 				LogstoreId:      "oss_config_id",
// 				Endpoint:        "endpoint",
// 				Project:         "project",
// 				Logstore:        "logstore",
// 				AccessKeyId:     "xxx",
// 				AccessKeySecret: "xxx",
// 			}}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetAclConfigDelta,
// 		func(*gorm.DB, string, string, string, int) (*model.DeltaObject, error) {
// 			return &model.DeltaObject{
// 				RegionNo:     "cn-hangzhou",
// 				RsPoolName:   "rsPoolName",
// 				Version:      1,
// 				BucketId:     "bucketId",
// 				Object:       "object",
// 				FirewallType: "dns",
// 			}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetAclConfigFull,
// 		func(*gorm.DB, string, string, string) (*model.DeltaObject, error) {
// 			return &model.DeltaObject{
// 				RegionNo:     "cn-hangzhou",
// 				RsPoolName:   "rsPoolName",
// 				Version:      1,
// 				BucketId:     "bucketId",
// 				Object:       "object",
// 				FirewallType: "dns",
// 			}, nil
// 		})

// 	patches.ApplyFunc(
// 		model.UpdateRsInfo,
// 		func(*gorm.DB, *model.Rs) error {
// 			return nil
// 		})

// 	patches.ApplyFunc(
// 		model.GetRegionSettings,
// 		func(*gorm.DB, string, string) ([]model.RegionSettings, error) {
// 			return []model.RegionSettings{{RegionNo: "cn-hangzhou", ConfigType: "IPS_GRAY", ConfigContent: "1"}}, nil
// 		})

// 	Convey("TestGetRsPoolBasicConfig", t, func() {
// 		w := httptest.NewRecorder()

// 		req := httptest.NewRequest("GET", "/api/firewall/config/getRsPoolBasicConfig?regionNo=cn-shanghai&rsPoolBiz=cfw&firewallIp=1.2.3.4", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)

// 		resp := HttpResp{}
// 		err := json.Unmarshal(w.Body.Bytes(), &resp)
// 		So(err, ShouldEqual, nil)
// 		So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 	})

// 	Convey("TestGetUserFullConfig", t, func() {
// 		w := httptest.NewRecorder()

// 		req := httptest.NewRequest("GET", "/api/firewall/config/getUserFullConfig?regionNo=cn-shanghai&rsPoolBiz=cfw&firewallIp=1.2.3.4", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)

// 		resp := HttpResp{}
// 		err := json.Unmarshal(w.Body.Bytes(), &resp)
// 		So(err, ShouldEqual, nil)
// 		So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 	})

// 	Convey("TestGetUserDeltaConfig", t, func() {
// 		w := httptest.NewRecorder()

// 		req := httptest.NewRequest("GET", "/api/firewall/config/getUserDeltaConfig?regionNo=cn-shanghai&rsPoolBiz=cfw&firewallIp=1.2.3.4&needFullConfig=false&currentVersion=2&workVersion=1&configType=acl_rule", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)

// 		resp := HttpResp{}
// 		err := json.Unmarshal(w.Body.Bytes(), &resp)
// 		So(err, ShouldEqual, nil)
// 		So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 	})

// 	Convey("TestGetUserDeltaConfigNeedFullConfig", t, func() {
// 		w := httptest.NewRecorder()

// 		req := httptest.NewRequest("GET", "/api/firewall/config/getUserDeltaConfig?regionNo=cn-shanghai&rsPoolBiz=cfw&firewallIp=1.2.3.4&needFullConfig=true&configType=acl_rule", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)

// 		resp := HttpResp{}
// 		err := json.Unmarshal(w.Body.Bytes(), &resp)
// 		So(err, ShouldEqual, nil)
// 		So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 	})

// 	Convey("TestPostRsInfo", t, func() {
// 		w := httptest.NewRecorder()

// 		req := httptest.NewRequest("GET", "/api/firewall/postRsInfo?regionNo=cn-shanghai&rsPoolBiz=cfw&firewallIp=1.2.3.4&rsPoolName=name&hostIp=1.2.3.4&firewallVersion=1&vswitchVersion=2&agentVersion=3", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)

// 		resp := HttpResp{}
// 		err := json.Unmarshal(w.Body.Bytes(), &resp)
// 		So(err, ShouldEqual, nil)
// 		So(resp.Msg, ShouldEqual, errcode.StatusSuccess.Msg)
// 	})
// }

// func TestGinRecovery(t *testing.T) {
// 	Convey("TestRecovery", t, func() {
// 		w := httptest.NewRecorder()
// 		req := httptest.NewRequest("GET", "/ops/rspool/describe?regionNo=cn-hangzhou", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "127.0.0.1")

// 		err := net.OpError{}

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				panic(&err)
// 			})
// 		defer patches.Reset()

// 		r := setupRouter()
// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, 500)
// 	})
// }

// func TestAuth(t *testing.T) {
// 	Convey("Test Auth Success", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			global.CONFIG.Region, bid, time.Now().Unix(), salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				global.CONFIG.Region, bid, time.Now().Unix(), sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "salt", bid, nil
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return sign, nil
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusOK)
// 	})

// 	Convey("Test Auth Fail: No Params", t, func() {
// 		r := setupRouter()

// 		w := httptest.NewRecorder()
// 		req := httptest.NewRequest("GET", "/ops/rspool/describe", nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusUnauthorized)
// 	})

// 	Convey("Test Auth Fail: GetAuthInfo Fail", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			global.CONFIG.Region, bid, time.Now().Unix(), salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				global.CONFIG.Region, bid, time.Now().Unix(), sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "salt", bid, errors.New("err")
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return sign, nil
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusUnauthorized)
// 	})

// 	Convey("Test Auth Fail: Salt Err", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			global.CONFIG.Region, bid, time.Now().Unix(), salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				global.CONFIG.Region, bid, time.Now().Unix(), sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "", bid, nil
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return sign, nil
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusUnauthorized)
// 	})

// 	Convey("Test Auth Fail: Bid Err", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			global.CONFIG.Region, bid, time.Now().Unix(), salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				global.CONFIG.Region, bid, time.Now().Unix(), sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "salt", "", nil
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return sign, nil
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusUnauthorized)
// 	})

// 	Convey("Test Auth Fail: QTime Err", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			global.CONFIG.Region, bid, 0, salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				global.CONFIG.Region, bid, 0, sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "salt", "bid", nil
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return sign, nil
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusUnauthorized)
// 	})

// 	Convey("Test Auth Fail: CalcSign err", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			global.CONFIG.Region, bid, time.Now().Unix(), salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				global.CONFIG.Region, bid, time.Now().Unix(), sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "salt", "bid", nil
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return sign, errors.New("err")
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusUnauthorized)
// 	})

// 	Convey("Test Auth Fail: sign err", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			global.CONFIG.Region, bid, time.Now().Unix(), salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				global.CONFIG.Region, bid, time.Now().Unix(), sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "salt", "bid", nil
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return "errSignString", nil
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusUnauthorized)
// 	})

// 	Convey("Test Auth Fail: Region err", t, func() {
// 		r := setupRouter()
// 		bid := "bid"
// 		salt := "salt"
// 		w := httptest.NewRecorder()
// 		region := "NoSuchRegion"
// 		signStr := fmt.Sprintf("regionNo=%s&identity=x&bid=%s&qtime=%d&salt=%s",
// 			region, bid, time.Now().Unix(), salt)
// 		sign, _ := utils.CalcSign(signStr)
// 		req := httptest.NewRequest("GET",
// 			fmt.Sprintf("/ops/rspool/describe?regionNo=%s&identity=x&bid=%s&qtime=%d&sign=%s",
// 				region, bid, time.Now().Unix(), sign),
// 			nil)

// 		//header
// 		req.Header.Add("X-Real-IP", "1.2.3.4")

// 		patches := gomonkey.ApplyFunc(
// 			rspool.DescribeRspool,
// 			func(*gin.Context) {
// 				return
// 			})
// 		defer patches.Reset()
// 		patches.ApplyFunc(
// 			model.GetAuthInfo,
// 			func(*gorm.DB, string) (string, string, error) {
// 				return "salt", "bid", nil
// 			})
// 		patches.ApplyFunc(
// 			utils.CalcSign,
// 			func(string) (string, error) {
// 				return sign, nil
// 			})

// 		r.ServeHTTP(w, req)

// 		So(w.Code, ShouldEqual, http.StatusBadRequest)
// 	})
// }
