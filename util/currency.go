package util

// 所有支持货币的常量
const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

// 验证输入的货币种类是否支持
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}
