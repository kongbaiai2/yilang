package utils

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// check cidr is network
// min <= mask <= max and != 0, example:cidr = 10.219.219.0/24
// 0 < min <= max <= 32
func ValidateCIDRByRangeMask(cidr string, min, max int) bool {
	if max < min || min <= 0 || max > 32 {
		return false
	}
	var mask int
	_, network, err := net.ParseCIDR(cidr)
	if err == nil {
		mask, _ = network.Mask.Size()
	}
	if cidr != network.String() {
		return false
	}
	return err == nil && mask <= max && mask >= min && mask != 0
}
func ValidateCIDR(cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)

	if cidr != network.String() {
		return false
	}
	return err == nil
}

func CalcMd5(str string) ([]byte, error) {
	h := md5.New()
	_, err := h.Write([]byte(str))
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// CalcSign calc HighIPEnd sign
func CalcSign(str string) (string, error) {
	md5sum, err := CalcMd5(str)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5sum), nil
}

func GenUUID() string {
	return uuid.New().String()
}

func GenUUID32() string {
	uuidStr := uuid.New().String()
	return strings.ReplaceAll(uuidStr, "-", "")
}

var validate = validator.New()

func StructToMap(input interface{}) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(input)
	typ := reflect.TypeOf(input)

	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		fieldValue := val.Field(i)

		// Only process if it's a string or something that can be converted to string
		if fieldValue.Kind() == reflect.String {
			result[fieldName] = fieldValue.String()
		} else {
			result[fieldName] = fmt.Sprintf("%v", fieldValue.Interface())
		}
	}

	return result
}

func StructToMapIgnoreZero(input interface{}) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(input)

	// 处理指针类型
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 确保输入是结构体
	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 获取JSON标签并处理字段名
		jsonTag := field.Tag.Get("json")
		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			} else if parts[0] == "-" {
				continue // 忽略该字段
			}
		}

		// 跳过非导出字段或零值（slice类型除外）
		if field.PkgPath != "" {
			continue
		}
		if fieldValue.IsZero() && fieldValue.Kind() != reflect.Slice {
			continue
		}

		// 处理slice类型（转换为JSON字符串）
		if fieldValue.Kind() == reflect.Slice {
			if fieldValue.IsZero() {
				result[fieldName] = "[]" // 空slice显式设置为"[]"
			} else {
				if jsonData, err := json.Marshal(fieldValue.Interface()); err == nil {
					result[fieldName] = string(jsonData)
				} else {
					// 序列化失败时回退到原始值
					result[fieldName] = fmt.Sprintf("%v", fieldValue.Interface())
				}
			}
		} else {
			// 非slice类型按原逻辑处理
			if fieldValue.Kind() == reflect.String {
				result[fieldName] = fieldValue.String()
			} else {
				result[fieldName] = fmt.Sprintf("%v", fieldValue.Interface())
			}
		}
	}
	return result
}

func MergeMaps(map1, map2 map[string]string) map[string]string {
	merged := make(map[string]string)

	// Add entries from the first map
	for key, value := range map1 {
		merged[key] = value
	}

	// Add entries from the second map
	// If key already exists, it will be overwritten
	for key, value := range map2 {
		merged[key] = value
	}

	return merged
}

// BindParamsPtr 将 Gin 的 GET 参数绑定到传入的请求对象
func BindParamsPtr(c *gin.Context, req interface{}) error {
	value := reflect.ValueOf(req)
	typ1 := value.Type()
	if req == nil || typ1.Kind() != reflect.Ptr {
		return fmt.Errorf("the input object cannot be nil or non pointer")
	}

	switch c.Request.Method {
	case "POST", "PUT", "DELETE":
		if err := c.BindJSON(req); err != nil {
			return err
		}
	case "GET":
		query := c.Request.URL.Query()
		val := reflect.ValueOf(req).Elem()
		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldVal := val.Field(i)

			// 获取 JSON 标签名
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			paramKey := strings.Split(jsonTag, ",")[0]

			// 获取查询参数值
			paramValue := query.Get(paramKey)
			if paramValue == "" {
				continue
			}

			// 根据字段类型赋值
			if err := setFieldValue(fieldVal, paramValue); err != nil {
				return fmt.Errorf("字段 %s 赋值失败: %v", field.Name, err)
			}
		}
	default:
		return errors.New("unknown http method")
	}
	if err := validate.Struct(req); err != nil {
		return err
	}
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	// 处理基础类型和切片 ptr
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int32, reflect.Int, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Slice:
		elemType := field.Type().Elem()
		slice := reflect.MakeSlice(field.Type(), 0, 0)

		// 处理逗号分隔的字符串（如 "1,2,3"）
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			elem := reflect.New(elemType).Elem()

			// 处理 []*int 类型
			if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Int {
				intVal, err := strconv.Atoi(part)
				if err != nil {
					return fmt.Errorf("无法将 %q 转换为整数: %v", part, err)
				}
				elem.Set(reflect.ValueOf(&intVal))
			} else {
				// 处理其他类型（如 []string, []*string）
				if err := setFieldValue(elem, part); err != nil {
					return err
				}
			}

			slice = reflect.Append(slice, elem)
		}

		field.Set(slice)
	case reflect.Ptr:
		// 创建指针指向的新对象
		elemType := field.Type().Elem()
		elem := reflect.New(elemType).Elem()
		if err := setFieldValue(elem, value); err != nil {
			return err
		}
		field.Set(elem.Addr())
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Type())
	}

	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 递归处理所有键的首字母大写
func CapitalizeKeys(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// 处理对象类型
		newMap := make(map[string]interface{})
		for key, value := range v {
			capitalizedKey := capitalizeFirst(key)
			newMap[capitalizedKey] = CapitalizeKeys(value)
		}
		return newMap
	case []interface{}:
		// 处理数组类型
		newSlice := make([]interface{}, len(v))
		for i, elem := range v {
			newSlice[i] = CapitalizeKeys(elem)
		}
		return newSlice
	default:
		// 其他类型直接返回
		return v
	}
}

// 将字符串首字母大写
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func FuncName() string {
	pc, _, _, _ := runtime.Caller(1) // 参数 1 表示向上追溯 1 层调用栈
	fullFuncName := runtime.FuncForPC(pc).Name()

	// 去除包路径（保留最后的包名.函数名）
	parts := strings.Split(fullFuncName, ".")
	return parts[len(parts)-1]
}

// interval的时间重试，maxTime的时间超时
// interval := 30 * time.Second
// maxTime := 300 * time.Second
func CallTimeOut(interval, maxTime time.Duration, f func() error) (err error) {
	i := 1
	for {
		time.Sleep(interval)
		err = f()
		if err == nil {
			return err
		}

		if time.Duration(i)*interval >= maxTime {
			return fmt.Errorf("call time out, interval: %v, maxTime: %v, retry: %v", interval, maxTime, i)
		}
		i++
	}
}

func ScientificToIntegerStr(scientificStr string) (string, error) {
	// 解析科学计数法字符串为 big.Float
	f := new(big.Float)
	_, err := fmt.Sscan(scientificStr, f)
	if err != nil {
		return "", err
	}

	// 转换为 big.Int
	i := new(big.Int)
	i, _ = f.Int(i) // 使用 Float 的 Int 方法转换为 Int

	// 返回整数字符串
	return i.String(), nil
}

func GetSystemConfigIntValueFromDB(db *sql.DB, name string, defaultValue int) int {
	var value int
	sqlStr := "select configValue from system_spec_config where configName = ?"
	err := db.QueryRow(sqlStr, name).Scan(&value)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[SMS] getSystemConfigValueFromDB: %v, %v", sqlStr, err)
		}
		return defaultValue
	}
	return value
}

func GetRegionSystemConfigIntValueFromDB(db *sql.DB, name string, defaultValue int, regionId string) int {
	var value int
	sqlStr := "select configValue from region_system_spec_config where configName = ? and regionId = ?"
	err := db.QueryRow(sqlStr, name, regionId).Scan(&value)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[SMS] getSystemConfigValueFromDB: %v, %v", sqlStr, err)
		}
		return defaultValue
	}
	return value
}
func GetAzoneSystemConfigIntValueFromDB(db *sql.DB, name string, defaultValue int, regionId string, azoneId string) int {
	var value int
	sqlStr := "select configValue from azone_system_spec_config where configName = ? and regionId = ? and azoneId = ?"
	err := db.QueryRow(sqlStr, name, regionId, azoneId).Scan(&value)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[SMS] getSystemConfigValueFromDB: %v, %v", sqlStr, err)
		}
		return defaultValue
	}
	return value
}

// Execute once every [interval] seconds, retry count [maxRetry], total time [timeout]
func RetryWithTimeout(interval, timeout time.Duration, maxRetry int, f func() error) error {
	timeoutCh := time.After(timeout)
	return retryWithTimeout(interval, maxRetry, timeoutCh, f)
}
func retryWithTimeout(interval time.Duration, maxRetry int, timeoutCh <-chan time.Time, f func() error) error {

	err := f()
	if err == nil {
		return nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	i := 0

	for {
		select {
		case <-timeoutCh:
			fmt.Println("task timeout")
			return fmt.Errorf("task timeout")
		case <-ticker.C:
			i++
			log.Printf("retry: %d\n", i)
			err := f()
			if err == nil {
				return nil
			}
			if maxRetry <= i {
				return fmt.Errorf("task timeout")
			}
			ticker.Reset(interval)
		}
	}
}

// 默认10秒重试，最大重试5次，10分钟超时
func RetryWithTimeoutDefault(f func() error) error {
	internal := 10 * time.Second
	retry_num := 5
	timeout := 10 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RetryWithTimeoutForCtx(ctx, internal, retry_num, f)
}

//	ctx, cancel := context.WithTimeout(context.Background(), MAXTIME)
//
// defer cancel()
func RetryWithTimeoutForCtx(ctx context.Context, interval time.Duration, maxRetry int, f func() error) error {
	err := f()
	if err == nil {
		return nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	i := 0
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation canceled: %w", ctx.Err())

		case <-ticker.C:
			i++
			log.Printf("retry: %d\n", i)

			if err := ctx.Err(); err != nil {
				return err
			}
			err := f()
			if err == nil {
				return nil
			}

			if err := ctx.Err(); err != nil {
				return err
			}
			if maxRetry <= i {
				return fmt.Errorf("task timeout")
			}
			ticker.Reset(interval)
		}
	}
}

// ResolveToIPv4 解析域名，返回 不为空的 IPv4 地址
func ResolveToIPv4(domain string) ([]net.IP, error) {
	// 可选：设置超时（避免卡住）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", domain)
	if err != nil {
		return []net.IP{}, fmt.Errorf("failed to resolve %q to IPv4: %w", domain, err)
	}
	if len(ips) == 0 {
		return []net.IP{}, fmt.Errorf("no IPv4 address found for %q", domain)
	}

	// ips[0].String()
	return ips, nil
}
