package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)  //去你指定的文件夹找资料
	viper.SetConfigName("app") //文件名
	viper.SetConfigType("env") //文件类型
	viper.AutomaticEnv()       //如果环境变量里也有，优先用环境变量的

	err = viper.ReadInConfig() //真正去读 app.env 文件。如果文件不存在或者写错了，就报错 return 结束
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config) //把读到的扁平文本数据，按照第一块里写的说明书（mapstructure 标签），一件一件塞进 config 这个结构体对象里
	return                         //这里不管err = viper.Unmarshal(&config)有没有问题，反正最后倒要返回
}
