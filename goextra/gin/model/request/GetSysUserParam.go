package request

import (
	"github.com/go-playground/validator/v10"
)

type GetSysUserParam struct {
	Id uint `form:"id" json:"id" binding:"required"`
	Name string `form:"name" json:"name" binding:"required" validate:"min=5,max=10"`
}

func(p *GetSysUserParam) GetError (err validator.ValidationErrors) string {
	var errStr string = ""
	for _, item := range err {
		switch item.Tag() {
		case "required":
			errStr += item.StructNamespace()+"必填"
		}
	}
	return errStr
}
