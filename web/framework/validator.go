package framework

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//增则
const (
	numericRegexString = "^[-+]?[0-9]+(?:\\.[0-9]+)?$"
)

//增则compole
var (
	numericRegex = regexp.MustCompile(numericRegexString)
)

//验证方法map
var vaildateMethod = map[string] func(reflect.Value, []string) bool {
	"required" : required,
	"between" : between,
	"numeric" : numeric,
}

//字段错误记录结构体
type FieldErr struct {
	FieldName string
	Message string
}

//校验总结构体
type Validator struct {
	Errors []FieldErr
}

//创建一个校验总结构体
func NewValidator() *Validator {
	return &Validator{
		Errors: make([]FieldErr, 0),
	}
}

func(vs * Validator) Validate(Inputs interface{}, message map[string]string) {
	explodeRules := vs.parseStruct(Inputs)
	vt := reflect.TypeOf(Inputs)
	vv := reflect.ValueOf(Inputs)
	for i := 0; i < vt.NumField(); i++ {//检查所有v结构
		FieldName := vt.Field(i).Name//字段名称
		FieldKind := vv.Field(i).Kind()//字段名称
		switch FieldKind {
		case reflect.Struct:
			//获取下一层提示message
			var nextMessage = make(map[string]string)
			nextPrefix := FieldName+">"
			for mKey, mMes := range message {
				if strings.HasPrefix(mKey, nextPrefix) {
					nextMessage[strings.Replace(mKey, nextPrefix, "", 1)] = FieldName+"."+mMes
				}
			}
			vs.Validate(vv.Field(i).Interface(), nextMessage)
		default:
			if items, ok := explodeRules[FieldName]; ok {
				for _, em := range items {
					methodStr,params := vs.parseRuleItem(em)
					if method, ok := vaildateMethod[methodStr]; ok {
						if method(vv.Field(i), params) {
							vs.parseMessage(FieldName, methodStr, message)
						}
					}
				}
			}
		}
	}
}

//解析添加用户需要提示的错误消息
func (vs *Validator) parseMessage(FieldName, methodName string, message map[string]string) {
	errMsg := ""
	if msg, ok := message[FieldName+"."+methodName]; ok {
		errMsg = msg
	} else {
		errMsg = FieldName+"参数有误，请核实"
	}
	e := FieldErr{
		FieldName: FieldName,
		Message: errMsg,
	}
	vs.Errors = append(vs.Errors, e)
}

//解析传入要校验的结构体和对应的提示
func (vs *Validator) parseStruct(s interface{}) map[string][]string {
	vt := reflect.TypeOf(s)
	explodeRules := make(map[string][]string)
	for i := 0; i < vt.NumField(); i++ {
		explodeRules[vt.Field(i).Name] = strings.Split(vt.Field(i).Tag.Get("vaildate"), "|")
	}
	return explodeRules
}

//解析校验方法和对应参数
func (vs *Validator) parseRuleItem(rule string) (string, []string) {
	var method string
	var params []string
	if strings.Index(rule, ":") != -1 {
		ruleSlice := strings.Split(rule, ":")
		if strings.Index(rule, ",") != -1 {
			params = strings.Split(ruleSlice[1], ",")
		} else {
			params = append(params, ruleSlice[1])
		}
		method = ruleSlice[0]
	} else {
		method = rule
	}
	return method,params
}

//非空，非默认值 "" nil struct{} interface{}
func required(value reflect.Value, params []string) bool {
	switch value.Kind() {
	case reflect.String:
		if value.String() == "" {
			return true
		}
	case reflect.Map,reflect.Slice:
		if !value.IsNil() {
			return true
		}
	case reflect.Array,reflect.Struct:
		if !value.IsZero() {
			return true
		}
	}
	return false
}

//between:1,20
func between(value reflect.Value, params []string) bool {
	switch value.Kind() {
	case reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64:
		num1, _ :=  strconv.ParseInt(params[0], 10, 64)
		num2, _ :=  strconv.ParseInt(params[1], 10, 64)
		if value.Int() < num1 {
			return true
		}
		if value.Int() > num2 {
			return true
		}
	case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
		num1, _ :=  strconv.ParseInt(params[0], 10, 64)
		num2, _ :=  strconv.ParseInt(params[1], 10, 64)
		if value.Uint() < uint64(num1) {
			return true
		}
		if value.Uint() > uint64(num2) {
			return true
		}
	case reflect.Float32,reflect.Float64:
		num1, _ :=  strconv.ParseInt(params[0], 10, 64)
		num2, _ :=  strconv.ParseInt(params[1], 10, 64)
		if value.Float() < float64(num1) {
			return true
		}
		if value.Float() > float64(num2) {
			return true
		}
	case reflect.String:
		num1, _ :=  strconv.Atoi(params[0])
		num2, _ :=  strconv.Atoi(params[1])
		if len([]rune(value.String())) < num1 {
			return true
		}
		if len([]rune(value.String())) > num2 {
			return true
		}
	case reflect.Array,reflect.Slice:
		num1, _ :=  strconv.Atoi(params[0])
		num2, _ :=  strconv.Atoi(params[1])
		if value.Len() < num1 {
			return true
		}

		if value.Len() > num2 {
			return true
		}
	}
	return false
}

//max:50
func max(value reflect.Value, params []string)  {

}

//max:10
func min(value reflect.Value, params []string)  {

}

//numeric
func numeric(value reflect.Value, params []string) bool {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64:
		return false
	default:
		return !numericRegex.MatchString(value.String())
	}
}

