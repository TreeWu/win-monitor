package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var notifyMap = make(map[string]Notify)

func main() {
	viper.SetConfigType("yaml")
	viper.SetConfigFile("confiy.yaml")
	var conf Config
	err := viper.Unmarshal(&conf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(conf)

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
	engine.Static("/console", "./win_monitor/dist")

	log.Fatal(engine.Run(":80"))
}

type MonitorServer struct {
	conf         Config
	db           *gorm.DB
	wechatServer *WechatServer
}

func (m *MonitorServer) RegisterApi(e *gin.Engine) {
	apiGroup := e.Group("api")
	{
		apiGroup.POST("register", m.Register)
		apiGroup.POST("monitor", m.Monitor)
		apiGroup.POST("screenshot", m.Screenshot)
		apiGroup.GET("devices", m.Devices)
		apiGroup.GET("device/:hostId", m.HostInfo)
	}
}

func (m *MonitorServer) Register(c *gin.Context) {
	var reg Host
	err := c.ShouldBindJSON(&reg)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: "Bad Request"})
		return
	}
	var host HostModel
	if err = m.db.Model(&HostModel{}).Where("host_id = ?", reg.HostID).First(&host).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			host = HostModel{
				Hostname:          reg.Hostname,
				OS:                reg.OS,
				Platform:          reg.Platform,
				HostID:            reg.HostID,
				PlatformFamily:    reg.PlatformFamily,
				PlatformVersion:   reg.PlatformVersion,
				CustomName:        reg.Hostname,
				FirstRegisterTime: time.Now().UnixMilli(),
				NotifyPush:        false,
				Config:            m.conf.MonitorConfig.Json(),
			}
			if err = m.db.Create(&host).Error; err != nil {
				c.JSON(http.StatusOK, Response{Code: 500, Msg: "success", Data: ""})
				return
			}
		}
	}
	var resp RegistryResp
	json.Unmarshal([]byte(host.Config), &resp)
	c.JSON(http.StatusOK, Response{Code: 200, Msg: "success", Data: resp})
}

func (m *MonitorServer) Monitor(c *gin.Context) {

	var mon Monitor
	err := c.ShouldBindJSON(&mon)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400, Msg: "Bad Request",
		})
		return
	}
	var monitors []MonitorModel
	for _, item := range mon.Items {
		monitor := MonitorModel{
			HostId:   mon.HostId,
			BootTime: item.BootTime,
			Type:     item.Type,
			Time:     item.Time,
			Total:    item.Total,
			Used:     item.Used,
			Free:     item.Free,
			Per:      item.Per,
			Unit:     item.Unit,
			Name:     item.Name,
		}
		if monitor.Name == "" {
			monitor.Name = monitor.Type
		}
		monitors = append(monitors, monitor)
	}
	err = m.db.CreateInBatches(&monitors, 100).Error
	c.JSON(http.StatusOK, Response{Code: 200, Msg: "success"})
}

func (m *MonitorServer) Screenshot(c *gin.Context) {

	var screenshot HostScreenshot
	err := c.ShouldBindJSON(&screenshot)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Bad Request"})
		return
	}
	go func(screenshot HostScreenshot) {
		if screenshot.Distance > m.conf.MonitorConfig.ScreenshotUploadMinDistance {
			return
		}
		notify := Notify{
			HostId:   screenshot.HostId,
			HostName: "",
			Time:     screenshot.CaptureTime,
			Distance: screenshot.Distance,
		}
		var host HostModel
		if err := m.db.Model(&HostModel{}).Where("host_id = ?", screenshot.HostId).First(&host).Error; err != nil {
			log.Println(err)
			return
		}
		if host.NotifyPush {
			notify.HostName = host.Hostname
			m.wechatServer.SendTemplate(c, notify)
		}
	}(screenshot)

	err = m.db.Model(&HostScreenshot{}).Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"pre", "cur", "distance", "capture_time"})}).Create(&screenshot).Error
	c.JSON(http.StatusOK, Response{Code: 200, Msg: "success"})
}

func (m *MonitorServer) Devices(c *gin.Context) {

	var devices []Host
	if err := m.db.Model(&HostModel{}).Find(&devices).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": devices})
}

func (m *MonitorServer) HostInfo(c *gin.Context) {

	hostId := c.Param("hostId")

	var monitors []MonitorModel
	m.db.Model(&MonitorModel{}).Debug().Select("type", "name", "boot_time", "time", "total", "used", "free", "per").Where("host_id = ? ", hostId).Order("id desc").Limit(m.conf.ServerConfig.Line).Find(&monitors)
	var screenshot HostScreenshot
	m.db.Model(&HostScreenshot{}).Where("host_id = ?", hostId).First(&screenshot)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": map[string]interface{}{
		"monitors": monitors, "screenshot": screenshot},
	})
}
