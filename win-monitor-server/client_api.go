package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/http"
	"time"
)

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

	}
	console := e.Group("console")
	{
		console.GET("host", m.Hosts)
		console.GET("host/:hostId", m.HostMonitor)
		console.POST("host/conf", m.UpdateHost)
	}
}

// Register 设备注册
//
//	@Summary		设备注册
//	@Tags			客户端接口
//	@Description	设备注册
//	@Accept			json
//	@Produce		json
//	@Param			参数	body		Host						true	"参数"
//	@Success		200	{object}	Response{data=MonitorConf}	"成功"
//	@Router			/api/register  [post]
func (m *MonitorServer) Register(c *gin.Context) {
	var reg Host
	err := c.ShouldBindJSON(&reg)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: http.StatusBadRequest, Msg: "Bad Request"})
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
				Config: MonitorConf{
					MonitorEnable:                 m.conf.MonitorConfig.MonitorEnable,
					ScreenshotEnable:              m.conf.MonitorConfig.ScreenshotEnable,
					MonitorUploadInterval:         m.conf.MonitorConfig.MonitorUploadInterval,
					MonitorCollectInterval:        m.conf.MonitorConfig.MonitorCollectInterval,
					MaxMonitorSize:                m.conf.MonitorConfig.MaxMonitorSize,
					ScreenshotUploadMinDistance:   m.conf.MonitorConfig.ScreenshotUploadMinDistance,
					ScreenshotUploadIntervalCount: m.conf.MonitorConfig.ScreenshotUploadIntervalCount,
					ScreenshotIntervalTime:        m.conf.MonitorConfig.ScreenshotIntervalTime,
					ScreenshotUploadOriginImage:   m.conf.MonitorConfig.ScreenshotUploadOriginImage,
				},
			}
			if err = m.db.Create(&host).Error; err != nil {
				c.JSON(http.StatusInternalServerError, Response{Code: http.StatusInternalServerError})
				return
			}
		}
	}
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Msg: "success", Data: host.Config})
}

// Monitor 监控数据上传
//
//	@Summary		监控数据上传
//	@Tags			客户端接口
//	@Description	监控数据上传
//	@Accept			json
//	@Produce		json
//	@Param			参数	body		Monitor					true	"参数"
//	@Success		200	{object}	Response{data=string}	"成功"
//	@Router			/api/monitor  [post]
func (m *MonitorServer) Monitor(c *gin.Context) {

	var mon Monitor
	err := c.ShouldBindJSON(&mon)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: http.StatusBadRequest, Msg: "Bad Request",
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
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Msg: "success"})
}

// Screenshot 截图上传
//
//	@Summary		截图上传
//	@Tags			客户端接口
//	@Description	设备注册
//	@Accept			json
//	@Produce		json
//	@Param			参数	body		HostScreenshot			true	"参数"
//	@Success		200	{object}	Response{data=string}	"成功"
//	@Router			/api/screenshot  [post]
func (m *MonitorServer) Screenshot(c *gin.Context) {

	var screenshot HostScreenshot
	err := c.ShouldBindJSON(&screenshot)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: http.StatusBadRequest, Msg: "Bad Request"})
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
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Msg: "success"})
}

// Hosts 主机列表
//
//	@Summary		主机列表
//	@Tags			控制台接口
//	@Description	主机列表
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Response{data=[]HostModel}	"成功"
//	@Router			/console/host  [get]
func (m *MonitorServer) Hosts(c *gin.Context) {

	var devices []HostModel
	if err := m.db.Model(&HostModel{}).Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: http.StatusInternalServerError, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Data: devices})
}

// HostMonitor 监控数据
//
//	@Summary		监控数据
//	@Tags			控制台接口
//	@Description	监控数据
//	@Accept			json
//	@Produce		json
//	@Param			hostId	path		string						true	"参数"
//	@Success		200		{object}	Response{data=HostMonitor}	"成功"
//	@Router			/console/host/:hostId  [get]
func (m *MonitorServer) HostMonitor(c *gin.Context) {
	hostId := c.Param("hostId")
	var monitors []MonitorModel
	m.db.Model(&MonitorModel{}).Debug().Select("type", "name", "boot_time", "time", "total", "used", "free", "per").Where("host_id = ? ", hostId).Order("id desc").Limit(m.conf.ServerConfig.Line).Find(&monitors)
	var screenshot HostScreenshot
	m.db.Model(&HostScreenshot{}).Where("host_id = ?", hostId).First(&screenshot)
	c.JSON(http.StatusOK, Response{Code: http.StatusOK, Data: HostMonitor{
		Monitors:   monitors,
		Screenshot: screenshot,
	},
	})
}

// UpdateHost 程序主机配置
//
//	@Summary		程序主机配置
//	@Tags			控制台接口
//	@Description	程序主机配置
//	@Accept			json
//	@Produce		json
//	@Param			参数	body		HostModel				true	"参数"
//	@Success		200	{object}	Response{data=string}	"成功"
//	@Router			/console/host/conf  [post]
func (m *MonitorServer) UpdateHost(c *gin.Context) {
	var host HostModel
	err := c.ShouldBindJSON(&host)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: http.StatusBadRequest, Msg: "Bad Request"})
		return
	}
	err = m.db.Model(&host).Where("host_id = ?", host.HostID).Updates(HostModel{HostID: host.HostID, CustomName: host.CustomName, Config: host.Config}).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: http.StatusBadRequest, Msg: "Bad Request"})
		return
	}
	c.JSON(http.StatusBadRequest, Response{Code: http.StatusOK})
}
