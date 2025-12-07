package cacti

import (
	"errors"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"

	"github.com/gin-gonic/gin"
)

type GetPercentMonthlyRequest struct {
	GraphID  int  `json:"GraphID" validate:"required"`
	MonthAgo int  `json:"MonthAgo" default:"1"`
	IsDown   bool `json:"IsDown"`
}

// Validate check request validation.
func (obj *GetPercentMonthlyRequest) Validate() *errcode.Err {

	return nil
}

type GetPercentMonthlyResponse struct {
	GetPercentEveryDayResponse
}

func GetPercentMonthlyWork(opt *GetPercentMonthlyRequest) (e *errcode.Err, ret *GetPercentEveryDayResponse) {

	monthStr, data, err := ProcessMonthly(opt.GraphID, opt.MonthAgo, opt.IsDown)
	if err != nil {
		return &errcode.Err{Msg: err.Error()}, nil
	}
	e = errcode.StatusSuccess
	ret = &GetPercentEveryDayResponse{
		Values: []DataValueResult{
			{
				Data:  monthStr,
				Value: data,
			},
		},
	}

	global.LOG.Infof("success, month p95 \n%v: %.2f 95th ", monthStr, data)

	return
}

func GetPercentMonthly(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
		opt := GetPercentMonthlyRequest{}
		if err := ginplus.BindParams(c, &opt); err != nil {
			global.LOG.Errorf("[ERROR] GetPercentMonthly check parameters failed, err:%v", err)
			if errors.Is(err, errcode.ErrorCidrFormat) {
				return errcode.ErrorCidrFormat, nil
			}
			return errcode.ErrorParameters, nil
		}

		e, ret = GetPercentMonthlyWork(&opt)
		// ret = listImages(c, ret_tmp)
		return
	})
}
