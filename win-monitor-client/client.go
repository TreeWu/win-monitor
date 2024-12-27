package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"time"

	"github.com/corona10/goimagehash"
	"github.com/disintegration/imaging"
	"github.com/kbinani/screenshot"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	apiHost         = "http://localhost"
	registerUrl     = apiHost + "api/client/register"
	monitorUrl      = apiHost + "api/client/monitor"
	screenshotUrl   = apiHost + "api/client/screenshot"
	registerTimeout = time.Second * 10
)

var httpClient = &http.Client{}
var MonitorChan = make(chan MonitorItem, 10000)

type Client struct {
	HostId string
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(ctx context.Context) *Client {
	return &Client{
		ctx: ctx,
	}
}

func (c *Client) Start() {

	ctx, cancel := context.WithCancel(c.ctx)
	c.cancel = cancel

	resp := c.Init(ctx)
	config := resp.Data

	if config.MonitorCollectInterval == 0 {
		config.MonitorCollectInterval = 30
	}
	if config.MonitorUploadInterval == 0 {
		config.MonitorUploadInterval = 60
	}
	if config.MaxMonitorSize == 0 {
		config.MaxMonitorSize = 100
	}
	if config.ScreenshotUploadMinDistance == 0 {
		config.ScreenshotUploadMinDistance = 5
	}
	if config.ScreenshotUploadIntervalCount == 0 {
		config.ScreenshotUploadIntervalCount = 5
	}
	if config.ScreenshotIntervalTime == 0 {
		config.ScreenshotIntervalTime = 30
	}
	if config.ScreenshotEnable {
		go c.DifferentScreen(ctx, config)
	}
	if config.MonitorEnable {
		go c.Collect(ctx, config)
		go c.Upload(ctx, config)
	}
}

func (c *Client) Init(ctx context.Context) Response {
	resp := Response{}
	for {
		info, err := host.Info()
		if err != nil {
			time.Sleep(registerTimeout)
			continue
		}
		bs, err := Post(ctx, registerUrl, Host{
			Hostname:        info.Hostname,
			OS:              info.OS,
			Platform:        info.Platform,
			HostID:          info.HostID,
			PlatformFamily:  info.PlatformFamily,
			PlatformVersion: info.PlatformVersion,
		})
		if err != nil {
			time.Sleep(registerTimeout)
			continue
		}
		err = json.Unmarshal(bs, &resp)
		if err != nil {
			time.Sleep(registerTimeout)
			continue
		}
		if resp.Code == 200 {
			c.HostId = info.HostID
			break
		}
	}
	return resp
}

func (c *Client) Upload(ctx context.Context, resp RegistryResp) {
	ticker := time.NewTicker(time.Duration(resp.MonitorUploadInterval) * time.Second)
	defer ticker.Stop()
	id, _ := host.HostID()
	monitor := Monitor{
		HostId: id,
	}
	for {
		select {
		case <-ticker.C:
			if len(monitor.Items) != 0 {
				_, err := Post(context.Background(), monitorUrl, monitor)
				if err == nil {
					monitor.Items = monitor.Items[0:0]
				}
			}
		case item := <-MonitorChan:
			monitor.Items = append(monitor.Items, item)
			if len(monitor.Items) >= resp.MaxMonitorSize {
				_, err := Post(context.Background(), monitorUrl, monitor)
				if err == nil {
					monitor.Items = monitor.Items[0:0]
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Collect(ctx context.Context, resp RegistryResp) {
	ticker := time.NewTicker(time.Duration(resp.MonitorCollectInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			info, err := host.Info()
			if err != nil {
				continue
			}
			if stat, err := cpu.Info(); err == nil {

				// 获取 CPU 使用率
				cpuPercent, err := cpu.Percent(time.Second, false)
				if err == nil {
					MonitorChan <- MonitorItem{
						Type:     MonitorTypeCPU,
						BootTime: int64(info.BootTime),
						Time:     time.Now().UnixMilli(),
						Per:      cpuPercent[0],
						Name:     MonitorTypeCPU + "[" + stat[0].ModelName + "]",
					}
				}
			}

			// 获取内存使用率
			memInfo, err := mem.VirtualMemory()
			if err == nil {
				MonitorChan <- MonitorItem{
					Type:     MonitorTypeMem,
					BootTime: int64(info.BootTime),
					Time:     time.Now().UnixMilli(),
					Total:    float64(memInfo.Total),
					Used:     float64(memInfo.Used),
					Free:     float64(memInfo.Available),
					Per:      memInfo.UsedPercent,
				}
			}

			if partitions, err := disk.Partitions(true); err == nil {
				for _, partition := range partitions {
					usage, err := disk.Usage(partition.Mountpoint)
					if err == nil {
						MonitorChan <- MonitorItem{
							Type:     MonitorTypeDisk,
							BootTime: int64(info.BootTime),
							Time:     time.Now().UnixMilli(),
							Total:    float64(usage.Total),
							Used:     float64(usage.Used),
							Free:     float64(usage.Free),
							Per:      usage.UsedPercent,
							Name:     MonitorTypeDisk + "[" + partition.Mountpoint + "]",
						}
					}
				}
			}

			// 这里可以添加上传状态的代码
			temperature, err := c.Temperature()
			if err == nil {
				MonitorChan <- MonitorItem{
					Type:     MonitorTypeTemperature,
					BootTime: int64(info.BootTime),
					Time:     time.Now().UnixMilli(),
					Used:     temperature,
				}
			}
		case <-ctx.Done():
			fmt.Println("Exiting...")
			return
		}
	}
}

func Post(ctx context.Context, url string, body interface{}) ([]byte, error) {
	marshal, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshal))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Client) DifferentScreen(ctx context.Context, resp RegistryResp) {
	var prevImg image.Image
	var prevHash *goimagehash.ImageHash
	var count int
	hostScreenshot := HostScreenshot{
		HostId:      c.HostId,
		Cur:         "",
		Pre:         "",
		Distance:    0,
		CaptureTime: 0,
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:

		}
		// 截取屏幕图片
		img, err := captureScreen()
		if err != nil {
			fmt.Println("截取屏幕失败:", err)
			continue
		}

		if prevImg != nil {
			// 计算当前图片的哈希值
			currHash, err := goimagehash.DifferenceHash(img)
			if err != nil {
				fmt.Println("计算哈希失败:", err)
				continue
			}

			// 比较当前图片与之前图片的哈希值
			hostScreenshot.Distance, err = prevHash.Distance(currHash)
			if err != nil {
				fmt.Println("计算哈希距离失败:", err)
				continue
			}
			// 变化度小于阈值或者连续上传次数达到阈值时，上传图片
			if hostScreenshot.Distance < resp.ScreenshotUploadMinDistance || count > resp.ScreenshotUploadIntervalCount {
				hostScreenshot.Cur, _ = imageToBase64(imageResize(img))
				hostScreenshot.Pre, _ = imageToBase64(imageResize(prevImg))
				hostScreenshot.CaptureTime = time.Now().UnixMilli()
				Post(ctx, screenshotUrl, hostScreenshot)
				count = 0
			}
			count++
		}
		prevImg = img
		prevHash, _ = goimagehash.DifferenceHash(img)
		time.Sleep(time.Duration(resp.ScreenshotIntervalTime) * time.Second)
	}
}

func (c *Client) Temperature() (float64, error) {
	temperatures, err := host.SensorsTemperatures()
	if err != nil {
		return 0, err
	}
	for _, sensor := range temperatures {
		if sensor.Temperature > 0 {
			return sensor.Temperature, nil
		}
	}
	return 0, errors.New("不支持")
}

func captureScreen() (image.Image, error) {
	// 获取屏幕尺寸
	bounds := screenshot.GetDisplayBounds(0)

	// 截取屏幕
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	if err != nil {
		return "", err
	}
	imgBytes := buf.Bytes()
	imgBase64 := base64.StdEncoding.EncodeToString(imgBytes)
	return imgBase64, nil
}

func imageResize(img image.Image) image.Image {
	return imaging.Resize(img, 600, 0, imaging.Lanczos)
}
