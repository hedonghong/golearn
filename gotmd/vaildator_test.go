package gotmd

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"testing"
)

func TestVaildare(t *testing.T) {
	type Child struct {
		Name string `vaildate:"required"`
	}
	type Person struct {
		Name string `vaildate:"required"`
		Children Child `vaildate:"required"`
		Age string `vaildate:"required"`
	}
	p := Person {
		Children: Child{},
	}
	message := make(map[string]string)
	message["Name.required"] = "Name必填，不能为空"
	message["Age.required"] = "Age必填，不能为空"
	message["Children.required"] = "xxx"
	message["Children>Name.required"] = "孩子Name必填，不能为空"
	vo := NewValidator()
	vo.Validate(p, message)
	fmt.Println(vo.Errors)
}

func TestVailatorV10(t *testing.T) {
	var validate = validator.New()
	type User struct {
		Name       string `validate:"required,min=1,max=16"`
		NickName   string `validate:"required,min=0,max=16"`
		Age        int    `validate:"required,gte=0,lte=130"`
		Email      string `validate:"required,email"`
		PassWord   string `validate:"required,min=6,max=16"`
		RepeatPass string `validate:"eqfield=Password"`
	}
	var user = User{
		Name:       "",
		NickName:   "B",
		Age:        10,
		Email:      "1@qq.com",
		PassWord:   "123456789",
		RepeatPass: "123456798",
	}
	errs := validate.Struct(user)
	if errs != nil {
		fmt.Println(errs.Error())
	}
}

//var ii int
//v := reflect.ValueOf(ii)
//p1 := hasValue(v) //false
//下面这个方式可以
//var ii *int
//ii = new(int)
//*ii = 0
//v := reflect.ValueOf(ii)
//p1 := hasValue(v) //true
//p1=p1
func hasValue(fl reflect.Value) bool {
	switch fl.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !fl.IsNil()
	default:
		p := fl.IsValid() // true
		p=p
		p1 := fl.Interface() // 0
		p1=p1
		p2 := fl.Type()
		p2=p2
		p3 := reflect.Zero(fl.Type())
		p3=p3
		p4:=reflect.Zero(fl.Type()).Interface() // 0
		p4=p4
		return fl.IsValid() && fl.Interface() != reflect.Zero(fl.Type()).Interface() // false
	}
}
