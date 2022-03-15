package util

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"golearn/goextra/gin/global"
	"strconv"
)

type JWT struct {
	//声明签名密钥
	SigningKey []byte
}

// 初始化jwt对象
func NewJWT() *JWT  {
	return &JWT{
		[]byte(global.MINI_CONFIG.Jwt.SigningKey),
	}
}

// 自定义有效载荷(这里采用自定义的UnionId作为有效载荷的一部分)
type CustomClaims struct {
	Id string
	Name string
	jwt.StandardClaims
}

func (c *CustomClaims) getId() uint {
	i, _ := strconv.Atoi(c.Id)
	return uint(i)
}

func (c *CustomClaims) getName() string {
	return c.Name
}

// 调用jwt-go库生成token
// 指定编码的算法为jwt.SigningMethodHS256
func (j *JWT) CreateToken (claims CustomClaims) (string, error)  {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

func (j *JWT) ParserToken(tokenString string) (*CustomClaims, error)  {
	// 输入用户自定义的Claims结构体对象,token,以及自定义函数来解析token字符串为jwt的Token结构体指针
	// Keyfunc是匿名函数类型: type Keyfunc func(*Token) (interface{}, error)
	// func ParseWithClaims(tokenString string, claims Claims, keyFunc Keyfunc) (*Token, error) {}
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	if err != nil {
		// jwt.ValidationError 是一个无效token的错误结构
		if ve, ok := err.(*jwt.ValidationError); ok {
			// ValidationErrorMalformed是一个uint常量，表示token不可用
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("token不可用")
				// ValidationErrorExpired表示Token过期
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token过期")
				// ValidationErrorNotValidYet表示无效token
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, errors.New("无效的token")
			} else {
				return nil, errors.New("token不可用")
			}
		}
	}

	// 将token中的claims信息解析出来并断言成用户自定义的有效载荷结构
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token无效")
}