package model

import (
	"time"
)

type SysUser struct {
	Id uint `json:"id" gorm:"column:id"` // 主键id
	Name string `json:"name" gorm:"column:name"` // 姓名
	Age uint8 `json:"age" gorm:"column:age"` // 年龄
	Phone string `json:"phone" gorm:"column:phone"` // 电话
	Email string `json:"email" gorm:"column:email"` // '邮箱
	Password string `json:"password" gorm:"column:password"` // 密码
	LoginedAt time.Time `gorm:"column:logined_at"`  // 最近登陆时间
	CreatedAt time.Time `gorm:"column:created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at"` // 更新时间
}
