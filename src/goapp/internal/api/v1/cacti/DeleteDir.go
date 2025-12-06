package cacti

import (
	"os"
	"path/filepath"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"

	"github.com/gin-gonic/gin"
)

// type DeleteDirRequest struct {
// 	GraphID  int    `json:"GraphID" validate:"required"`
// 	Procent  string `json:"Percent" default:"95"`
// 	MonthAgo int    `json:"MonthAgo" default:"1"`
// 	IsDown   bool   `json:"IsDown"`
// }

// // Validate check request validation.
// func (obj *DeleteDirRequest) Validate() *errcode.Err {

// 	return nil
// }

type DeleteDirResponse struct {
	GetPercentEveryDayResponse
}

// func DeleteDirWork(opt *DeleteDirRequest) (e *errcode.Err, ret *GetPercentEveryDayResponse) {

// 	monthStr, data, err := ProcessMonthly(opt.GraphID, opt.MonthAgo, opt.IsDown)
// 	if err != nil {
// 		return &errcode.Err{Msg: err.Error()}, nil
// 	}
// 	e = errcode.StatusSuccess
// 	ret = &GetPercentEveryDayResponse{
// 		values: []DataValueResult{
// 			{
// 				Data:  monthStr,
// 				Value: data / 1000000,
// 			},
// 		},
// 	}

// 	global.LOG.Errorf("success, month p95 \n%v: %.2f 95th ", monthStr, data/1000000)

// 	return
// }

func DeleteDir(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {

		dir := global.CONFIG.CactiCfg.ImgPath
		err := os.RemoveAll(filepath.Clean(dir))
		if err != nil {
			global.LOG.Errorf("[ERROR] DeleteDir remove directory failed, err:%v", err)
			return &errcode.Err{Msg: err.Error()}, nil
		}
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			global.LOG.Errorf("[ERROR] DeleteDir recreate directory failed, err:%v", err)
			return &errcode.Err{Msg: err.Error()}, nil
		}

		e = errcode.StatusSuccess
		return
	})
}
