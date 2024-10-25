package main

const (
	MonitorTypeCPU         = "cpu"
	MonitorTypeMem         = "mem"
	MonitorTypeDisk        = "disk"
	MonitorTypeTemperature = "temperature"
	MonitorTypeOpen        = "open"
)

type RegistryResp struct {
	MonitorEnable                 bool `json:"monitorEnable"`                 // 监控开关
	ScreenshotEnable              bool `json:"screenshotEnable"`              // 截图开关
	MonitorUploadInterval         int  `json:"monitorUploadInterval"`         // 监控上传间隔时间
	MonitorCollectInterval        int  `json:"monitorCollectInterval"`        // 采集间隔时间
	MaxMonitorSize                int  `json:"maxMonitorSize"`                // 监控数据最大保存数量
	ScreenshotUploadMinDistance   int  `json:"screenshotUploadMinDistance"`   // 当前截图和上次截图对比，如果相似度小于该值则上传
	ScreenshotUploadIntervalCount int  `json:"screenshotUploadIntervalCount"` // 多少张截图后强制上传
	ScreenshotIntervalTime        int  `json:"screenshotIntervalTime"`        // 截图间隔时间
	ScreenshotUploadOriginImage   bool `json:"screenshotUploadOriginImage"`   // 是否上传原图
}

type Host struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`              // ex: freebsd, linux
	Platform        string `json:"platform"`        // ex: ubuntu, linuxmint
	HostID          string `json:"hostID"`          // ex: uuid
	PlatformFamily  string `json:"platformFamily"`  // ex: debian, rhel
	PlatformVersion string `json:"platformVersion"` // version of the complete OS
}

type Monitor struct {
	HostId string `json:"hostId"`

	Items []MonitorItem `json:"items"`
}

type MonitorItem struct {
	Type     string  `json:"type,omitempty"`
	BootTime int64   `json:"bootTime,omitempty"`
	Time     int64   `json:"time,omitempty"`
	Total    float64 `json:"total,omitempty"`
	Used     float64 `json:"used,omitempty"`
	Free     float64 `json:"free,omitempty"`
	Per      float64 `json:"per,omitempty"`
	Unit     string  `json:"unit,omitempty"`
	Name     string  `json:"name,omitempty"`
}

type HostModel struct {
	Id                int    `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT" json:"id"`
	Hostname          string `gorm:"column:hostname;type:varchar(64)"`
	OS                string `gorm:"column:os;type:varchar(64)"`
	Platform          string `gorm:"column:platform;type:varchar(64)"`
	HostID            string `gorm:"column:host_id;type:varchar(64);unique:host_id"`
	PlatformFamily    string `gorm:"column:platform_family;type:varchar(64)"`
	PlatformVersion   string `gorm:"column:platform_version;type:varchar(64)"`
	CustomName        string `gorm:"column:custom_name;type:varchar(64)"`
	FirstRegisterTime int64  `gorm:"column:first_register_time;type:bigint"`
	NotifyPush        bool   `gorm:"column:notify_push"`
	Config            string `gorm:"column:config;type:longtext;"`
}

func (h *HostModel) TableName() string {
	return "host"
}

type MonitorModel struct {
	Id       int     `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT" json:"id,omitempty"`
	HostId   string  `gorm:"column:host_id;type:varchar(64);index" json:"host_id,omitempty"` // 主机唯一标识
	Type     string  `gorm:"column:type;type:varchar(32)" json:"type,omitempty"`             // 监控类型 cpu/mem/disk/open
	BootTime int64   `gorm:"column:boot_time;type:bigint;index" json:"boot_time,omitempty"`  // 开机时间
	Time     int64   `gorm:"column:time;type:bigint" json:"time,omitempty"`                  // 监控时间
	Total    float64 `gorm:"column:total;type:double" json:"total"`                          // 总量
	Used     float64 `gorm:"column:used;type:double" json:"used"`                            // 已使用
	Free     float64 `gorm:"column:free;type:double" json:"free"`                            // 空闲
	Per      float64 `gorm:"column:per;type:double" json:"per"`                              // 使用率
	Unit     string  `gorm:"column:unit;type:varchar(32)" json:"unit,omitempty"`             // 单位
	Name     string  `gorm:"column:name;type:varchar(64)" json:"name,omitempty"`             // 名称
}

func (m *MonitorModel) TableName() string {
	return "monitor"

}

type HostScreenshot struct {
	Id          int    `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT" json:"id,omitempty"`
	HostId      string `gorm:"column:host_id;type:varchar(64);uniqueIndex" json:"host_id,omitempty"` // 主机唯一标识
	Cur         string `json:"cur" gorm:"type:longtext;"`
	Pre         string `json:"pre" gorm:"type:longtext;"`
	Distance    int    `json:"distance"`
	CaptureTime int64  `json:"captureTime"`
}

func (h *HostScreenshot) TableName() string {
	return "host_screenshot"
}

type Notify struct {
	HostName     string
	Time         int64
	Distance     int
	LastDistance int
	LastTime     int64
	HostId       string
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
