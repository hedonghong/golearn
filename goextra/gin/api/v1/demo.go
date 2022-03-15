package v1

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golearn/goextra/gin/global"
	"golearn/goextra/gin/model"
	"golearn/goextra/gin/model/request"
	"golearn/goextra/gin/model/response"
	"golearn/goextra/gin/util"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func GetSysUser(c *gin.Context)  {
	var param request.GetSysUserParam
	err := c.ShouldBindQuery(&param)
	if err != nil {
		fmt.Println(fmt.Errorf("%s \n", param.GetError(err.(validator.ValidationErrors))))
		return
	}
}

func PostSysUser(c *gin.Context)  {
	var param request.GetSysUserParam
	err := c.ShouldBindJSON(&param)
	if err != nil {
		fmt.Println(fmt.Errorf("%s \n", param.GetError(err.(validator.ValidationErrors))))
		return
	}
	validate := validator.New()
	err = validate.Struct(param)
	var user *model.SysUser = &model.SysUser{}
	errOrm := global.MINI_DB.Where("id = ?", 1).First(user).Error
	if errOrm == gorm.ErrRecordNotFound {
		fmt.Println(errOrm.Error)
		return
	} else if errOrm != nil {
		fmt.Println(err.Error)
		return
	}
	jsonUser , _ := json.Marshal(user)
	fmt.Println(string(jsonUser))
}

func UserLogin(c *gin.Context)  {
	var userLogin request.UserLoginParam
	err := c.ShouldBindJSON(&userLogin)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	var user *model.SysUser = &model.SysUser{}
	errOrm := global.MINI_DB.Where("name = ?", userLogin.Name).First(user).Error
	if errOrm == gorm.ErrRecordNotFound {
		response.FailWithMessage("请输入正确的用户/密码", c)
		return
	} else if errOrm != nil {
		response.FailWithMessage(errOrm.Error(), c)
		return
	}
	if userLogin.Password != user.Password {
		response.FailWithMessage("请输入正确的用户/密码", c)
		return
	}
	userId := string(strconv.Itoa(int(user.Id)))
	// 下发token
	j := util.NewJWT()
	//180秒过期
	sec := time.Duration(86400)
	standClaim := jwt.StandardClaims {
		Audience: "gin_test",
		ExpiresAt: time.Now().Add(time.Second * sec).Unix(),
		Id: userId,
		IssuedAt:time.Now().Unix(),
		Issuer:"gin_test",
		Subject:"gin_test",
	}
	token, _ := j.CreateToken(util.CustomClaims{
		Id: userId,
		Name: user.Name,
		StandardClaims: standClaim,
	})
	response.OkWithData(token, c)
	return
}
