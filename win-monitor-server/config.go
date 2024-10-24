package main

type Config struct {
	Port int
}

type WatcherConfig struct {
	ScreenshotInterval          int      `yaml:"screenshotInterval"`
	ScreenshotUploadMinDistance int      `yaml:"screenshotUploadMinDistance"`
	MessagePushInterval         int      `yaml:"messagePushInterval"`
	ScreenshotIntervalTime      int      `yaml:"screenshotIntervalTime"`
	ScreenshotPath              string   `yaml:"screenshotPath"`
	WechatServerUrl             string   `yaml:"wechatServerUrl"`
	WechatServerSecret          string   `yaml:"wechatServerSecret"`
	WechatServerToken           string   `yaml:"wechatServerToken"`
	WechatServerAppId           string   `yaml:"wechatServerAppId"`
	WechatServerAppSecret       string   `yaml:"wechatServerAppSecret"`
	WechatServerTemplateId      string   `yaml:"wechatServerTemplateId"`
	WechatServerTemplateData    string   `yaml:"wechatServerTemplateData"`
	WechatServerTemplateUrl     string   `yaml:"wechatServerTemplateUrl"`
	ToUser                      []string `yaml:"toUser"`
}
