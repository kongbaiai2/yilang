package cidr

import (
	"errors"

	"github.com/kongbaiai2/yilang/goapp/pkg/utils"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"
	"github.com/kongbaiai2/yilang/goapp/pkg/model"

	"github.com/gin-gonic/gin"
)

type DeleteCidrRequest struct {
	Cidr string `json:"cidr" validate:"required"`
}

// Validate check request validation.
func (obj *DeleteCidrRequest) Validate() *errcode.Err {
	var maskMin = 16
	var maskMax = 24
	if !utils.ValidateCIDRByRangeMask(obj.Cidr, maskMin, maskMax) {
		return errcode.ErrorCidrFormat
	}
	return nil
}

type DeleteCidrResponse struct {
	Cidr string
}

func DeleteCidrWork(opt *DeleteCidrRequest) (e *errcode.Err, ret *DeleteCidrResponse) {
	err := model.DeleteQueryId(global.DB, opt.Cidr)
	if err != nil {
		global.LOG.Errorf("[ERROR] DeleteQueryId failed, err:%v", err)
		return errcode.ErrorDBDelete, nil
	}
	e = errcode.StatusSuccess
	ret = &DeleteCidrResponse{Cidr: opt.Cidr}

	return
}

func DeleteCidr(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
		opt := DeleteCidrRequest{}
		if err := ginplus.BindParams(c, &opt); err != nil {
			global.LOG.Errorf("[ERROR] DeleteCidr check parameters failed, err:%v", err)
			if errors.Is(err, errcode.ErrorCidrFormat) {
				return errcode.ErrorCidrFormat, nil
			}
			return errcode.ErrorParameters, nil
		}

		e, ret = DeleteCidrWork(&opt)
		return
	})
}
