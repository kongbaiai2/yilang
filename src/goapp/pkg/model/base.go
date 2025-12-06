package model

import "time"

type Base struct {
	ID        uint64    `json:"id,omitempty" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseFull struct {
	ID        uint64    `json:"id,omitempty" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// INode define here
type INode struct {
	Name        string    `json:"name" gorm:"column:name;type:varchar(256);not null" validate:"required"`
	Dir         string    `json:"dir" gorm:"column:dir;type:varchar(1000);not null" validate:"required"`
	Fullname    string    `json:"fullname" gorm:"column:fullname;type:varchar(1000);not null" validate:"required"`
	LocalPath   string    `json:"local_path" gorm:"-"`
	Size        uint64    `json:"size" gorm:"column:size;type:int(10)"`
	Type        string    `json:"type" gorm:"column:type;type:enum('file','directory','symlink','device','special')"`
	Linkto      string    `json:"linkto,omitempty" gorm:"column:linkto"`
	Md5sum      string    `json:"md5sum,omitempty" gorm:"column:md5sum;type:varchar(64)"`
	Crc64ECMA   string    `json:"crc64ecma,omitempty" gorm:"-"`
	IModifyTime time.Time `json:"inode_modify_time" gorm:"column:inode_modify_time"`
	IDeleted    bool      `json:"inode_deleted" gorm:"column:inode_deleted"`
}
