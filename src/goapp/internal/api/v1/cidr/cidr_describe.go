package cidr

import (
	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"
	"github.com/kongbaiai2/yilang/goapp/pkg/model"

	"github.com/gin-gonic/gin"
)

type DescribeCidrRequest struct {
	Name string `json:"name"`
	Cidr string `json:"cidr"`
}
type ListCidrsRequest struct {
	DescribeCidrRequest
}

// Validate check request validation.
func (obj *DescribeCidrRequest) Validate() error {

	return nil
}

type DescribeCidrResponse struct {
	Detail []model.NfvCidrResponse `json:"detail"`
}

func CreateListCidrsRequest() *DescribeCidrRequest {
	return &DescribeCidrRequest{}
}

func ListCidrsWork(opt *DescribeCidrRequest) (e *errcode.Err, ret *DescribeCidrResponse) {
	importResp, err := model.GetAllNfvCidr(global.DB)
	if err != nil {
		global.LOG.Errorf("[ERROR] ListCidrsWork failed, err:%v", err)
		return errcode.ErrorOpenApiError, nil
	}
	e = errcode.StatusSuccess
	ret = &DescribeCidrResponse{Detail: importResp}
	return
}

func ListCidrs(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
		opt := DescribeCidrRequest{}
		if err := ginplus.BindParams(c, &opt); err != nil {
			global.LOG.Errorf("[ERROR] ListCidrs check parameters failed, err:%v", err)
			return errcode.ErrorParameters, nil
		}
		// global.LOG.Infof("%v", opt)
		e, ret = ListCidrsWork(&opt)
		return
	})
}
