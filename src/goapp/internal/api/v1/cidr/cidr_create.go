package cidr

// import (
// )

// type CreateCidrRequest struct {
// 	Name   string `json:"name" validate:"required"`
// 	Cidr   string `json:"cidr" validate:"required"`
// 	AliUid string `json:"aliuid" validate:"required"`
// }

// var (
// 	// limit mask range
// 	maskMin = 16
// 	maskMax = 24
// )

// // // Validate check request validation.
// // func (obj *CreateCidrRequest) Validate() *errcode.Err {
// // 	if !utils.ValidateCIDRByRangeMask(obj.Cidr, maskMin, maskMax) {
// // 		return errcode.ErrorCidrFormat
// // 	}
// // 	return nil
// // }

// type CreateCidrResponse struct {
// 	Cidr string `json:"cidr"`
// }

// type LbListener struct {
// 	FrontendPort int              `json:"frontend_port"`
// 	BackendPort  int              `json:"backend_port"`
// 	VipProtocol  string           `json:"vip_protocol"`
// 	Schedule     string           `json:"schedule"`
// 	Check        zeus.HealthCheck `json:"check"`
// }

// func CreateCidrWork(opt *CreateCidrRequest) (e *errcode.Err, ret *CreateCidrResponse) {

// 	// 查询是否已创建
// 	isExist, err := model.GetNfvCidrByCidr(global.DB, opt.Cidr)
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		global.LOG.Errorf("[ERROR] GetNfvCidrByCidr failed, err:%v", err)
// 		return errcode.ErrorDBSelect, nil
// 	}
// 	if err == nil && isExist.Cidr != "" {
// 		return errcode.ErrorCidrExist, nil
// 	}

// 	request := model.CreateNfvCidrObjectRequest()
// 	request.Cidr = opt.Cidr
// 	request.Name = opt.Name
// 	request.Status = 0
// 	request.AliUid = opt.AliUid
// 	request.GmtCreate = time.Now()
// 	request.GmtModify = time.Now()
// 	global.LOG.Infof("%v", request)
// 	if err := model.CreateNfvCidr(global.DB, request); err != nil {
// 		global.LOG.Errorf("[ERROR] CreateNfvCidr save db failed, cidr: %+v, err:%v", opt.Cidr, err)
// 		return errcode.ErrorDBInsert, nil
// 	}

// 	e = errcode.StatusSuccess
// 	ret = &CreateCidrResponse{Cidr: opt.Cidr}

// 	return
// }

// func CreateCidr(c *gin.Context) {
// 	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
// 		opt := CreateCidrRequest{}
// 		if err := ginplus.BindParams(c, &opt); err != nil {
// 			global.LOG.Errorf("[ERROR] CreateCidr check parameters failed, err:%v", err)
// 			if errors.Is(err, errcode.ErrorCidrFormat) {
// 				return errcode.ErrorCidrFormat, nil
// 			}
// 			return &errcode.Err{Code: errcode.ErrorParameters.Code, Msg: err.Error()}, nil
// 		}
// 		global.LOG.Infof("%v", opt)

// 		e, ret = CreateCidrWork(&opt)
// 		return
// 	})
// }
