package cacti

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"

	"github.com/gin-gonic/gin"
)

// type DescribeImageRequest struct {
// 	GraphID  int    `json:"GraphID" validate:"required"`
// 	Procent  string `json:"Percent" default:"95"`
// 	MonthAgo int    `json:"MonthAgo" default:"1"`
// 	IsDown   bool   `json:"IsDown"`
// }

// // Validate check request validation.
// func (obj *DescribeImageRequest) Validate() *errcode.Err {

// 	return nil
// }

type DescribeImageResponse struct {
	GetPercentEveryDayResponse
}

// func DescribeImageWork(opt *DescribeImageRequest) (e *errcode.Err, ret *GetPercentEveryDayResponse) {

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

func DescribeImage(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
		// opt := DescribeImageRequest{}
		// if err := ginplus.BindParams(c, &opt); err != nil {
		// 	global.LOG.Errorf("[ERROR] DescribeImage check parameters failed, err:%v", err)
		// 	if errors.Is(err, errcode.ErrorCidrFormat) {
		// 		return errcode.ErrorCidrFormat, nil
		// 	}
		// 	return errcode.ErrorParameters, nil
		// }

		dir := global.CONFIG.CactiCfg.ImgPath
		files, err := os.ReadDir(dir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取目录"})
			return
		}

		var images []map[string]string
		for _, file := range files {
			if !file.IsDir() && (filepath.Ext(file.Name()) == ".png" || filepath.Ext(file.Name()) == ".jpg") {
				// 构造图片的相对路径
				imagePath := dir + "/" + file.Name()
				images = append(images, map[string]string{
					"name": file.Name(),
					"url":  imagePath,
				})
			}
		}

		// 返回JSON格式的数据，包括图片列表和一条文本信息

		ret = map[string]interface{}{
			"images": images,
			"url":    "http://" + global.CONFIG.System.IpAddress + global.CONFIG.System.HttpPort + "/",
		}
		return
	})
}
