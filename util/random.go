package util

import (
	"math/rand"
	"strings"
	"time"
)

const alpahet = "abcdefghijklmnopqrstuvwxyz"

// 在程序启动时，给随机数生成器设置一个种子，让每次例如rand.Intn(100)的值不一样
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 随机产生一个整形的数,介于min和max之间
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1) //min->max-min+min

}

// 随机产生长度为n的字符串
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alpahet)
	for i := 0; i < n; i++ {
		c := alpahet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()

}

// 随机产生长度为6的人名
func RandomOwner() string {
	return RandomString(6)
}

// 随机产生0-1000的钱
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// 随机产生货币种类
func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)]

}
