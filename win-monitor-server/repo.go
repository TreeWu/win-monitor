package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func NewMysql(conf MysqlConfig) *gorm.DB {

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: conf.Username + ":" + conf.Password + "@tcp(" + conf.Host + ":" + conf.Port + ")/" + conf.Database + "?charset=utf8&parseTime=True&loc=Local",
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&HostModel{}, &MonitorModel{}, HostScreenshot{})
	if err != nil {
		log.Fatal(err)
	}
	return db
}
