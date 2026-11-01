package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/techschool/simplebank/util"
)

// 验证函数就是一个：专门接收框架传过来的待测字段数据（FieldLevel），执行完判断逻辑后，向框架报告最终结论（true 或 false）的标准化函数。
var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok { //从带着binding:"required,currency "的字段中获取字段值
		return util.IsSupportedCurrency(currency) //获取成功，检查是否满足IsSupportedCurrency函数的要求
	}
	return false
}
