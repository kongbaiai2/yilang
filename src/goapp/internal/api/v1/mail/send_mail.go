package mail

import (
	"crypto/tls"
	"errors"

	"github.com/kongbaiai2/yilang/goapp/internal/global"
	"github.com/kongbaiai2/yilang/goapp/pkg/errcode"
	"github.com/kongbaiai2/yilang/goapp/pkg/ginplus"

	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
)

type SendMailRequest struct {
	Smtp   SmtpConfig `json:"smtp"`
	Header MailHeader `json:"header"`
}

type SmtpConfig struct {
	Host     string `json:"Host" validate:"required"`
	Port     int    `json:"Port" validate:"required"`
	Username string `json:"Username" validate:"required,email"`
	Password string `json:"Password" validate:"required"`
}

type MailHeader struct {
	From    string   `json:"From" validate:"required,email"`
	To      []string `json:"To" validate:"required,dive,email"`
	Subject string   `json:"Subject" validate:"required"`
	Body    string   `json:"Body" validate:"required"`
	Attach  []string `json:"Attach"`
}

// Validate check request validation.
func (obj *SendMailRequest) Validate() *errcode.Err {

	return nil
}

func SendMailWork(opt *SendMailRequest) (e *errcode.Err) {
	m := gomail.NewMessage()
	m.SetHeader("From", opt.Header.From)
	m.SetHeader("To", opt.Header.To...)
	m.SetHeader("Subject", opt.Header.Subject)
	m.SetBody("text/html", opt.Header.Body)
	for _, file := range opt.Header.Attach {
		m.Attach(file)
	}

	d := gomail.NewDialer(opt.Smtp.Host, opt.Smtp.Port, opt.Smtp.Username, opt.Smtp.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		global.LOG.Errorf("[ERROR] SendMail send mail failed, err:%v", err)
		return &errcode.Err{Msg: err.Error()}
	}

	e = errcode.StatusSuccess
	return
}

func SendMail(c *gin.Context) {
	ginplus.ResponseWrapper(c, func(c *gin.Context) (e *errcode.Err, ret interface{}) {
		opt := SendMailRequest{}
		if err := ginplus.BindParams(c, &opt); err != nil {
			global.LOG.Errorf("[ERROR] SendMail check parameters failed, err:%v", err)
			if errors.Is(err, errcode.ErrorCidrFormat) {
				return errcode.ErrorCidrFormat, nil
			}
			return errcode.ErrorParameters, nil
		}

		global.LOG.Debugf("[INFO] SendMail parameters: %+v", opt)
		e = SendMailWork(&opt)
		// ret = listImages(c, ret_tmp)
		return
	})
}
