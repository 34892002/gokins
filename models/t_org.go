package models

import (
	"time"
)

type TOrg struct {
	Id          string    `xorm:"not null pk VARCHAR(64)" json:"id"`
	Aid         int64     `xorm:"not null pk autoincr BIGINT(20)" json:"aid"`
	Uid         string    `xorm:"VARCHAR(64)" json:"uid"`
	Name        string    `xorm:"VARCHAR(100)" json:"name"`
	Desc        string    `xorm:"TEXT" json:"desc"`
	Public      int       `xorm:"default 0 comment('公开') INT(1)" json:"public"`
	Created     time.Time `xorm:"comment('创建时间') DATETIME" json:"created"`
	Updated     time.Time `xorm:"comment('更新时间') DATETIME" json:"updated"`
	Deleted     int       `xorm:"default 0 INT(1)" json:"-"`
	DeletedTime time.Time `xorm:"DATETIME" json:"-"`
}

type TOrgInfo struct {
	Id          string    `xorm:"not null pk VARCHAR(64)" json:"id"`
	Aid         int64     `xorm:"not null pk autoincr BIGINT(20)" json:"aid"`
	Uid         string    `xorm:"VARCHAR(64)" json:"uid"`
	Name        string    `xorm:"VARCHAR(100)" json:"name"`
	Desc        string    `xorm:"TEXT" json:"desc"`
	Public      int       `xorm:"default 0 comment('公开') INT(1)" json:"public"`
	Created     time.Time `xorm:"comment('创建时间') DATETIME" json:"created"`
	Updated     time.Time `xorm:"comment('更新时间') DATETIME" json:"updated"`
	Deleted     int       `xorm:"default 0 INT(1)" json:"-"`
	DeletedTime time.Time `xorm:"DATETIME" json:"-"`

	PermAdm  int `xorm:"default 0 comment('管理员') INT(1)" json:"permAdm"`
	PermRw   int `xorm:"default 0 comment('1只读,2读写') INT(1)" json:"permRw"`
	PermExec int `xorm:"default 0 comment('执行权限') INT(1)" json:"permExec"`
}

func (TOrgInfo) TableName() string {
	return "t_org"
}
