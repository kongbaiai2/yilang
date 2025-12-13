package cacti

import (
	"errors"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"

	"github.com/gin-gonic/gin"
)

type GetPercentEveryDayRequest struct {
	GraphID  int  `json:"GraphID" validate:"required"`
	MonthAgo int  `json:"MonthAgo" default:"1"`
	IsDown   bool `json:"IsDown"`
}

// Validate check request validation.
func (obj *GetPercentEveryDayRequest) Validate() *errcode.Err {

	return nil
}

type GetPercentEveryDayResponse struct {
	Values []DataValueResult
	Url    string
}

func GetPercentEveryDayWork(opt *GetPercentEveryDayRequest) (e *errcode.Err, ret *GetPercentEveryDayResponse) {

	data, err := ProcessDaily(opt.GraphID, opt.MonthAgo, opt.IsDown)
	if err != nil {
		return &errcode.Err{Msg: err.Error()}, nil
	}
	e = errcode.StatusSuccess
	ret = &GetPercentEveryDayResponse{Values: data}
	// for _, v := range data {
	// 	global.LOG.Infof("success, day p95: \n%v: %.2f ", v.Data, v.Value)
	// }
	// global.LOG.Infof("success, day p95: %v ", data)

	return
}

func GetPercentEveryDay(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
		opt := GetPercentEveryDayRequest{}
		if err := ginplus.BindParams(c, &opt); err != nil {
			global.LOG.Errorf("[ERROR] GetPercentEveryDay check parameters failed, err:%v", err)
			if errors.Is(err, errcode.ErrorCidrFormat) {
				return errcode.ErrorCidrFormat, nil
			}
			return errcode.ErrorParameters, nil
		}

		global.LOG.Infof("%+v", opt)
		e, ret = GetPercentEveryDayWork(&opt)
		// ret = listImages(c, ret_tmp)
		return
	})
}
