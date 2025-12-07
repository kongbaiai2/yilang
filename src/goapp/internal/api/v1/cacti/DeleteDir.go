package cacti

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"

	"github.com/gin-gonic/gin"
)

type DeleteDirRequest struct {
	Dir string `json:"Dir"`
}

// // Validate check request validation.
// func (obj *DeleteDirRequest) Validate() *errcode.Err {

// 	return nil
// }

func DeleteDirWork(opt *DeleteDirRequest) (e *errcode.Err, ret interface{}) {
	dir := opt.Dir
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
}

func DeleteDir(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
		opt := DeleteDirRequest{}
		if err := ginplus.BindParams(c, &opt); err != nil {
			global.LOG.Errorf("[ERROR] GetPercentEveryDay check parameters failed, err:%v", err)
			if errors.Is(err, errcode.ErrorCidrFormat) {
				return errcode.ErrorCidrFormat, nil
			}
			return errcode.ErrorParameters, nil
		}

		if opt.Dir == "" {
			opt.Dir = global.CONFIG.CactiCfg.ImgPath
		}
		return DeleteDirWork(&opt)
	})
}
