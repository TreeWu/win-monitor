package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const (
	LINE                        = 4000
	ScreenshotUploadMinDistance = 1
	ScreenshotIntervalTime      = 30
	MessagePushInterval         = time.Minute * 5
	toUser                      = "oVpRZ6vCVcjDTTE9NNVpvWtqi1Zs"
	templateID                  = "280hiUk950o8NNNFi8L0P2XWJagLdC0X3JfHm8kIWKw"
	templateUrl                 = "https://{{host}}/console/host/{{hostId}}"
	appID                       = "wx96343687b5b4a2cd"
	appSecret                   = "3e1cd392c18d446be6b302a918fdd05d"
	token                       = "shrimp"
)

var notifyMap = make(map[string]Notify)

func main() {
	Server(context.Background())
}

func Server(ctx context.Context) {

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: "root:123456@tcp({{host}}:3306)/niugexi?charset=utf8&parseTime=True&loc=Local",
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&HostModel{}, &MonitorModel{}, HostScreenshot{})
	if err != nil {
		log.Fatal(err)
	}

	engine := gin.Default()
	engine.Use(gzip.Gzip(gzip.BestCompression))
	server := newWechatServer()

	apiGroup := engine.Group("api")
	{
		apiGroup.POST("register", func(c *gin.Context) {
			var reg Host
			err := c.ShouldBindJSON(&reg)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Bad Request"})
				return
			}
			host := HostModel{
				Hostname:          reg.Hostname,
				OS:                reg.OS,
				Platform:          reg.Platform,
				HostID:            reg.HostID,
				PlatformFamily:    reg.PlatformFamily,
				PlatformVersion:   reg.PlatformVersion,
				CustomName:        reg.Hostname,
				FirstRegisterTime: time.Now().UnixMilli(),
				NotifyPush:        false,
			}
			err = db.Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"hostname", "os", "platform", "platform_family", "platform_version"})}).Create(&host).Error
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": RegistryResp{
				MonitorEnable:                 true,
				ScreenshotEnable:              true,
				MonitorUploadInterval:         120,
				MonitorCollectInterval:        60,
				MaxMonitorSize:                100,
				ScreenshotUploadMinDistance:   ScreenshotUploadMinDistance,
				ScreenshotUploadIntervalCount: 6,
				ScreenshotIntervalTime:        ScreenshotIntervalTime,
			}})
		})

		apiGroup.POST("monitor", func(c *gin.Context) {
			var mon Monitor
			err := c.ShouldBindJSON(&mon)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Bad Request"})
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
			err = db.CreateInBatches(&monitors, 100).Error
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
		})
		apiGroup.POST("screenshot", func(c *gin.Context) {
			var screenshot HostScreenshot
			err := c.ShouldBindJSON(&screenshot)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Bad Request"})
				return
			}
			go func(screenshot HostScreenshot) {
				if screenshot.Distance > ScreenshotUploadMinDistance {
					return
				}
				notify := Notify{
					HostId:   screenshot.HostId,
					HostName: "",
					Time:     screenshot.CaptureTime,
					Distance: screenshot.Distance,
				}
				var host HostModel
				if err := db.Model(&HostModel{}).Where("host_id = ?", screenshot.HostId).First(&host).Error; err != nil {
					log.Println(err)
					return
				}
				if host.NotifyPush {
					notify.HostName = host.Hostname
					server.SendTemplate(ctx, notify)
				}
			}(screenshot)

			err = db.Model(&HostScreenshot{}).Clauses(clause.OnConflict{DoUpdates: clause.AssignmentColumns([]string{"pre", "cur", "distance", "capture_time"})}).Create(&screenshot).Error
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
		})

		apiGroup.GET("devices", func(c *gin.Context) {
			var devices []Host
			if err := db.Model(&HostModel{}).Find(&devices).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 500, "msg": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": devices})
		})

		apiGroup.GET("device/:hostId", func(c *gin.Context) {
			hostId := c.Param("hostId")

			var monitors []MonitorModel
			db.Model(&MonitorModel{}).Debug().Select("type", "name", "boot_time", "time", "total", "used", "free", "per").Where("host_id = ? ", hostId).Order("id desc").Limit(LINE).Find(&monitors)
			var screenshot HostScreenshot
			db.Model(&HostScreenshot{}).Where("host_id = ?", hostId).First(&screenshot)
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": map[string]interface{}{
				"monitors": monitors, "screenshot": screenshot},
			})
		})
	}

	wechatGroup := engine.Group("wechat")
	{
		wechatGroup.GET("api", func(c *gin.Context) {
			server.GetServer(c.Request, c.Writer)
		})
		wechatGroup.POST("api", func(c *gin.Context) {
			server.GetServer(c.Request, c.Writer)
		})
		wechatGroup.GET("template/send", func(c *gin.Context) {
			server.SendTemplate(c, Notify{
				HostId:   "test",
				HostName: "test",
				Time:     time.Now().UnixMilli(),
				Distance: rand.Intn(10),
			})
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
		})
	}

	// 提供静态文件
	engine.Static("/static", "./win_monitor/dist")
	engine.Static("/dist", "./win_monitor/dist")

	// 提供index.html文件
	engine.NoRoute(func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	err = engine.Run(":80")
	if err != nil {
		return
	}
}

type wechatServer struct {
	officialAccount *officialaccount.OfficialAccount
}

func newWechatServer() *wechatServer {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()

	cfg := &config.Config{
		AppID:     appID,
		AppSecret: appSecret,
		Token:     token,
		Cache:     memory,
	}
	officialAccount := wc.GetOfficialAccount(cfg)
	return &wechatServer{officialAccount: officialAccount}
}
func (w *wechatServer) GetServer(req *http.Request, writer http.ResponseWriter) {

	srv := w.officialAccount.GetServer(req, writer)
	srv.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		text := message.NewText(msg.Content)
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
	})
	err := srv.Serve()
	if err != nil {
		fmt.Println(err)
		return
	}
	srv.Send()
}

func (w *wechatServer) SendTemplate(ctx context.Context, notify Notify) {
	var old Notify
	var ok bool
	if old, ok = notifyMap[notify.HostId]; ok {
		if time.UnixMilli(notify.Time).Sub(time.UnixMilli(old.LastTime)) < MessagePushInterval {
			return
		}
	} else {
		old.HostId = notify.HostId
		old.LastTime = notify.Time
		old.LastDistance = notify.LastDistance
	}
	template := message.NewTemplate(w.officialAccount.GetContext())
	_, err := template.Send(&message.TemplateMessage{
		ToUser:     toUser,
		TemplateID: templateID,
		URL:        templateUrl,
		Data: map[string]*message.TemplateDataItem{
			"hostname": {
				Value: notify.HostName,
			},
			"time": {
				Value: time.UnixMilli(notify.Time).Format(time.DateTime),
			},
			"distance": {
				Value: strconv.Itoa(notify.Distance),
			},
			"lastdistance": {
				Value: strconv.Itoa(old.LastDistance),
			},
			"lasttime": {
				Value: time.UnixMilli(old.LastTime).Format(time.DateTime),
			},
		},
	})
	if err != nil {
		log.Println("模版消息发送", err)
		return
	}
	old.LastDistance = notify.Distance
	old.LastTime = notify.Time
	notifyMap[notify.HostId] = old
}
