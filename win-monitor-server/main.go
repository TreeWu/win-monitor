package main

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
)

var notifyMap = make(map[string]Notify)

func main() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile("./conf.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	var conf Config
	err = viper.Unmarshal(&conf)
	if err != nil {
		log.Fatal(err)
	}
	db := NewMysql(conf.MysqlConfig)
	engine := gin.Default()
	engine.Use(gzip.Gzip(gzip.BestCompression))
	wechatServer := newWechatServer(conf.WechatConfig)
	wechatServer.RegisterApi(engine)
	monitorServer := MonitorServer{
		conf:         conf,
		db:           db,
		wechatServer: wechatServer,
	}
	monitorServer.RegisterApi(engine)
	// 提供静态文件
	engine.Static("/admin", "./resource/console")
	log.Fatal(engine.Run(":80"))
}
