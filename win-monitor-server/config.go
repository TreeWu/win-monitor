package main

import "encoding/json"

type Config struct {
	ServerConfig  ServerConfig  `yaml:"server"`
	WechatConfig  WechatConfig  `yaml:"wechat"`
	MonitorConfig MonitorConfig `yaml:"monitor"`
	MysqlConfig   MysqlConfig   `yaml:"mysql"`
}

type ServerConfig struct {
	Port int
	Line int
}
type MonitorConfig struct {
	MonitorEnable                 bool `yaml:"monitorEnable"`                 // 监控开关
	ScreenshotEnable              bool `yaml:"screenshotEnable"`              // 截图开关
	MonitorUploadInterval         int  `yaml:"monitorUploadInterval"`         // 监控上传间隔时间
	MonitorCollectInterval        int  `yaml:"monitorCollectInterval"`        // 采集间隔时间
	MaxMonitorSize                int  `yaml:"maxMonitorSize"`                // 监控数据最大保存数量
	ScreenshotUploadMinDistance   int  `yaml:"screenshotUploadMinDistance"`   // 当前截图和上次截图对比，如果相似度小于该值则上传
	ScreenshotUploadIntervalCount int  `yaml:"screenshotUploadIntervalCount"` // 多少张截图后强制上传
	ScreenshotIntervalTime        int  `yaml:"screenshotIntervalTime"`        // 截图间隔时间
	ScreenshotUploadOriginImage   bool `yaml:"screenshotUploadOriginImage"`   // 是否上传原图
}

func (m MonitorConfig) Json() string {
	s, _ := json.Marshal(m)
	return string(s)
}

type MysqlConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type WechatConfig struct {
	Token               string `yaml:"token"`
	AppId               string `yaml:"appId"`
	AppSecret           string `yaml:"appSecret"`
	TemplateId          string `yaml:"templateId"`
	TemplateUrl         string `yaml:"templateUrl"`
	ToUser              string `yaml:"toUser"`
	MessagePushInterval int    `yaml:"messagePushInterval"`
}
