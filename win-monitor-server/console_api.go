package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Hosts 主机列表
//
//	@Summary		主机列表
//	@Tags			控制台接口
//	@Description	主机列表
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Response{data=[]HostModel}	"成功"
//	@Router			/api/console/host  [get]
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
//	@Router			/api/console/host/:hostId  [get]
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
//	@Router			/api/console/host/conf  [post]
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
	c.JSON(http.StatusOK, Response{Code: http.StatusOK})
}
