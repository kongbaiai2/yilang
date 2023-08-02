package cmcc_api

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm/clause"
)

// CmccDomainAdd
func CmccDomainAdd(domain, isp, supplier string) error {
	ins := &CmccDomain{Domain: domain, Isp: isp, Supplier: supplier}
	ins.CreatedAt = time.Now()
	ins.UpdatedAt = time.Now()
	// return db.Create(ins).Error
	return db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&ins).Error
}

func CmccDomainAll(search string) ([]CmccDomain, error) {
	var hs []CmccDomain
	query := db.Order("updated_at")
	if search != "" {
		query = db.Where("name like ?", "%"+search+"%")
	}
	err := query.Find(&hs).Error
	return hs, err
}

func CmccDomainFind(idx uint) (*CmccDomain, error) {
	ins := &CmccDomain{}
	ins.ID = idx
	return ins, db.First(ins).Error
}

func CmccDomainDelete(idx uint) error {
	ins := CmccDomain{}
	ins.ID = idx
	return db.Where("id = ?", idx).Delete(&ins).Error
}
func CmccDomainDeleteAll() error {
	ins := CmccDomain{}
	return db.Delete(&ins).Error
}

func CmccDomainUpdate(domain, isp, supplier string, id uint) error {
	ins := CmccDomain{Domain: domain, Isp: isp, Supplier: supplier}
	wh := &CmccDomain{}
	wh.ID = id
	return db.Model(wh).Updates(ins).Error
}

func CmccDomainDuplicate(idx uint) error {
	ins := &CmccDomain{}
	ins.ID = idx
	err := db.First(ins).Error
	if err != nil {
		return err
	}
	ins.ID = 0
	ins.Domain = fmt.Sprintf("%s_du", ins.Domain)
	return db.Create(ins).Error
}

// Update a row
func (m *CmccDomain) Update() (err error) {
	return db.Model(m).Update("", m).Error
}

// CreateUserOfRole insert a row
func (m *CmccDomain) Create() (err error) {
	m.ID = 0
	return db.Create(m).Error
}

// Delete destroy a row
func (m *CmccDomain) Delete() (err error) {
	if m.ID == 0 {
		return errors.New("resource must not be zero value")
	}
	return crudDelete(m)
}

func (m *CmccDomain) ChangeUpdateTime() (err error) {
	m.UpdatedAt = time.Now()
	return db.Save(m).Error
}

func crudDelete(m interface{}) (err error) {
	//WARNING When delete a record, you need to ensure it’s primary field has value, and GORM will use the primary key to delete the record, if primary field’s blank, GORM will delete all records for the model
	//primary key must be not zero value
	db := db.Delete(m)
	if err = db.Error; err != nil {
		return
	}
	if db.RowsAffected != 1 {
		return errors.New("resource is not found to destroy")
	}
	return nil
}

// func crudOne(m interface{}, one interface{}) (err error) {
// 	if db.Where(m).First(one).RecordNotFound() {
// 		return errors.New("resource is not found")
// 	}
// 	return nil
// }
