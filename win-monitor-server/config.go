package main

import "encoding/json"

type Config struct {
	ServerConfig  ServerConfig  `mapstructure:"server"`
	WechatConfig  WechatConfig  `mapstructure:"wechat"`
	MonitorConfig MonitorConfig `mapstructure:"monitor"`
	MysqlConfig   MysqlConfig   `mapstructure:"mysql"`
}

type ServerConfig struct {
	Port int
	Line int
}
type MonitorConfig struct {
	MonitorEnable                 bool `mapstructure:"monitorEnable"`                 // 监控开关
	ScreenshotEnable              bool `mapstructure:"screenshotEnable"`              // 截图开关
	MonitorUploadInterval         int  `mapstructure:"monitorUploadInterval"`         // 监控上传间隔时间
	MonitorCollectInterval        int  `mapstructure:"monitorCollectInterval"`        // 采集间隔时间
	MaxMonitorSize                int  `mapstructure:"maxMonitorSize"`                // 监控数据最大保存数量
	ScreenshotUploadMinDistance   int  `mapstructure:"screenshotUploadMinDistance"`   // 当前截图和上次截图对比，如果相似度小于该值则上传
	ScreenshotUploadIntervalCount int  `mapstructure:"screenshotUploadIntervalCount"` // 多少张截图后强制上传
	ScreenshotIntervalTime        int  `mapstructure:"screenshotIntervalTime"`        // 截图间隔时间
	ScreenshotUploadOriginImage   bool `mapstructure:"screenshotUploadOriginImage"`   // 是否上传原图
}

func (m MonitorConfig) Json() string {
	s, _ := json.Marshal(m)
	return string(s)
}

type MysqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type WechatConfig struct {
	Token               string `mapstructure:"token"`
	AppId               string `mapstructure:"appId"`
	AppSecret           string `mapstructure:"appSecret"`
	TemplateId          string `mapstructure:"templateId"`
	TemplateUrl         string `mapstructure:"templateUrl"`
	ToUser              string `mapstructure:"toUser"`
	MessagePushInterval int    `mapstructure:"messagePushInterval"`
}
