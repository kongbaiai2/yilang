// Package ginplus generate by gCLI, don't edit it.
package ginplus

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"

	"github.com/gin-gonic/gin"
	validator "gopkg.in/go-playground/validator.v9"
)

var (
	errUnknownMethod = errors.New("unknown http method")

	// value validations for structs and individual fields based on tags.
	validate = validator.New()
)

// ResponseHandler handle *gin.Context.
type ResponseHandler func(c *gin.Context) (e *errcode.Err, ret interface{})

/*

type APIResponse struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Msg     string      `json:"message"`
	Module  string      `json:"module,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
*/

// ResponseWrapper warp response to json payload.
func ResponseWrapper(c *gin.Context, fn ResponseHandler) {
	err, ret := fn(c)
	// TODO yangchu 验证nil是否被强转类型
	if err == nil || err == errcode.StatusSuccess {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code":    200,
			"success": true,
			"message": "successful",
			"data":    ret,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"code":    err.Code,
		"success": false,
		"message": err.Error(),
		"data":    ret,
	})
	c.Set("cfw_code", err.Code)
}

// BindParams parse request params to a struct.
func BindParams(c *gin.Context, params interface{}) error {
	switch c.Request.Method {
	case "GET":
		val := reflect.ValueOf(params).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			if key := field.Tag.Get("json"); key != "" {
				if v := c.Query(key); v != "" {
					switch field.Type.String() {
					case "string":
						val.FieldByIndex(field.Index).SetString(v)
					case "int", "int32", "int64":
						n, err := strconv.ParseInt(v, 10, 64)
						if err != nil {
							return err
						}
						val.FieldByIndex(field.Index).SetInt(n)
					case "uint", "uint32", "uint64":
						n, err := strconv.ParseUint(v, 10, 64)
						if err != nil {
							return err
						}
						val.FieldByIndex(field.Index).SetUint(n)
					case "bool":
						n, err := strconv.ParseBool(v)
						if err != nil {
							return err
						}
						val.FieldByIndex(field.Index).SetBool(n)
					// case "ginplus.JSONUnixTime":
					// 	n, err := strconv.ParseInt(v, 10, 64)
					// 	if err != nil {
					// 		return err
					// 	}
					// 	val.FieldByIndex(field.Index).Set(reflect.ValueOf(JSONUnixTime{time.Unix(n, 0)}))
					default:
						return fmt.Errorf("unknown type: %v", field)
					}
				}
			}
		}
	case "POST", "PUT", "DELETE":
		if err := c.BindJSON(params); err != nil {
			return err
		}
	default:
		return errUnknownMethod
	}

	if err := bindDefaultValue(params); err != nil {
		return err
	}

	if err := validate.Struct(params); err != nil {
		return err
	}

	return callValidate(params)
}

// callValidate call method: `func (obj *Obj)Validate() error` and return result.
func callValidate(params interface{}) error {
	if method := reflect.ValueOf(params).MethodByName("Validate"); method.IsValid() {
		if ret := method.Call([]reflect.Value{}); ret != nil && len(ret) == 1 {
			if ret[0].IsNil() {
				return nil
			} else if e, ok := ret[0].Interface().(error); ok {
				return e
			}
		}
	}

	return nil
}

func bindDefaultValue(params interface{}) error {
	val := reflect.ValueOf(params).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.FieldByIndex(field.Index)
		if isEmptyValue(fieldValue) {
			defaultValue := field.Tag.Get("default")
			if defaultValue != "" {
				switch field.Type.String() {
				case "string":
					fieldValue.SetString(defaultValue)
				case "int", "int32", "int64":
					n, err := strconv.ParseInt(defaultValue, 10, 64)
					if err != nil {
						return err
					}
					fieldValue.SetInt(n)
				case "uint", "uint32", "uint64":
					n, err := strconv.ParseUint(defaultValue, 10, 64)
					if err != nil {
						return err
					}
					fieldValue.SetUint(n)
				// case "ginplus.JSONUnixTime":
				// 	n, err := strconv.ParseInt(defaultValue, 10, 64)
				// 	if err != nil {
				// 		return err
				// 	}
				// 	fieldValue.Set(reflect.ValueOf(JSONUnixTime{time.Unix(n, 0)}))
				default:
					return fmt.Errorf("unknown type: %v", field)
				}
			}
		}
	}

	return nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		if vt, ok := v.Interface().(JSONUnixTime); ok {
			return vt.Unix() == -62135596800
		}
	}
	return false
}

// JSONUnixTime wrap time.Time for JSON Marshal & Unmarshal.
type JSONUnixTime struct {
	time.Time
}

// MarshalJSON implements the json.Marshaler interface.
func (t JSONUnixTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%d", t.Unix())
	return []byte(stamp), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *JSONUnixTime) UnmarshalJSON(b []byte) error {
	n, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	*t = JSONUnixTime{time.Unix(n, 0)}
	return nil
}
