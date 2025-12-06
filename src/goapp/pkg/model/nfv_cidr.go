package model

import (
	"errors"
	"time"

	"github.com/kongbaiai2/yilang/goapp/internal/global"

	"gorm.io/gorm"
)

type NfvCidrObject struct {
	ID        uint64    `json:"id" gorm:"primarykey"`
	Name      string    `json:"name" gorm:"column:name"`
	Cidr      string    `json:"cidr" gorm:"column:cidr"  validate:"required"`
	Status    int       `json:"status" gorm:"column:status"`
	AliUid    string    `json:"aliuid" gorm:"column:aliuid"`
	GmtCreate time.Time `json:"gmt_create" gorm:"column:gmt_create"`
	GmtModify time.Time `json:"gmt_modify" gorm:"column:gmt_modify"`
}
type NfvCidrResponse struct {
	Name   string `json:"name" gorm:"column:name"`
	Cidr   string `json:"cidr" gorm:"column:cidr"  validate:"required"`
	Status int    `json:"status" gorm:"column:status"`
}

func (obj *NfvCidrObject) TableName() (name string) {
	name = "nfv_cidr"
	return
}
func (obj *NfvCidrResponse) TableName() (name string) {
	name = "nfv_cidr"
	return
}
func CreateNfvCidrObjectRequest() *NfvCidrObject {
	return &NfvCidrObject{}
}

func GetAllNfvCidr(db *gorm.DB) ([]NfvCidrResponse, error) {
	rs := []NfvCidrResponse{}
	if err := db.Find(&rs).Error; err != nil {
		global.LOG.Errorf("get all failed: %s", err)
		return nil, err
	}
	return rs, nil
}

func GetNfvCidrByCidr(db *gorm.DB, cidr string) (*NfvCidrObject, error) {
	var obj NfvCidrObject
	tx := db.Where("cidr = ? ", cidr)

	if err := tx.First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			global.LOG.Infof("uk %s record not found", cidr)
		} else {
			global.LOG.Errorf("%s find object by unique key %s failed: %s", obj.TableName(), cidr, err)
		}
		return &obj, err
	}
	return &obj, nil
}

func CreateNfvCidr(db *gorm.DB, nfvCidr *NfvCidrObject) error {
	return db.Create(nfvCidr).Error
}

func DeleteQueryId(db *gorm.DB, cidr string) error {
	return db.Where("cidr = ? ", cidr).Delete(&NfvCidrObject{}).Error
}
