package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WechatServer struct {
	officialAccount *officialaccount.OfficialAccount
	conf            WechatConfig
}

func newWechatServer(c WechatConfig) *WechatServer {
	wc := wechat.NewWechat()
	memory := cache.NewMemory()

	cfg := &config.Config{
		AppID:     c.AppId,
		AppSecret: c.AppSecret,
		Token:     c.Token,
		Cache:     memory,
	}
	officialAccount := wc.GetOfficialAccount(cfg)
	return &WechatServer{officialAccount: officialAccount, conf: c}
}
func (w *WechatServer) GetServer(req *http.Request, writer http.ResponseWriter) {

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

func (w *WechatServer) SendTemplate(ctx context.Context, notify Notify) {
	var old Notify
	var ok bool
	if old, ok = notifyMap[notify.HostId]; ok {
		if time.UnixMilli(notify.Time).Sub(time.UnixMilli(old.LastTime)) < time.Second*time.Duration(w.conf.MessagePushInterval) {
			return
		}
	} else {
		old.HostId = notify.HostId
		old.LastTime = notify.Time
		old.LastDistance = notify.LastDistance
	}
	template := message.NewTemplate(w.officialAccount.GetContext())
	_, err := template.Send(&message.TemplateMessage{
		ToUser:     w.conf.ToUser,
		TemplateID: w.conf.TemplateId,
		URL:        strings.ReplaceAll(w.conf.TemplateUrl, "{{hostId}}", notify.HostId),
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

func (w *WechatServer) RegisterApi(e *gin.Engine) {
	wechatGroup := e.Group("wechat")
	{
		wechatGroup.GET("api", func(c *gin.Context) {
			w.GetServer(c.Request, c.Writer)
		})
		wechatGroup.POST("api", func(c *gin.Context) {
			w.GetServer(c.Request, c.Writer)
		})
		wechatGroup.GET("template/send", func(c *gin.Context) {
			w.SendTemplate(c, Notify{
				HostId:   "test",
				HostName: "test",
				Time:     time.Now().UnixMilli(),
				Distance: rand.Intn(10),
			})
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success"})
		})
	}

}
