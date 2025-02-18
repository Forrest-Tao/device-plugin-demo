package device_plugin

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"io/fs"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
	"path/filepath"
	"strings"
)

type DeviceMonitor struct {
	path    string
	devices map[string]*pluginapi.Device
	notify  chan struct{} // notify when device update
}

func NewDeviceMonitor(path string) *DeviceMonitor {
	return &DeviceMonitor{
		path:    path,
		devices: make(map[string]*pluginapi.Device),
		notify:  make(chan struct{}),
	}
}

// List 将 /etc/xpu 下所有设备保存到 map中
func (d *DeviceMonitor) List() error {
	err := filepath.Walk(d.path, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			klog.Info("%s is dir,skip", path)
			return nil
		}

		d.devices[info.Name()] = &pluginapi.Device{
			ID:       info.Name(),
			Health:   pluginapi.Healthy,
			Topology: nil,
		}
		return nil
	})
	if err == nil {
		return nil
	}
	return errors.WithMessagef(err, "walk [%s] failed", d.path)
}

// Watch 监视 /etc/xpu 下所有设备变更
func (d *DeviceMonitor) Watch(watcher *fsnotify.Watcher) error {
	klog.Infoln("watching devices")
	if err := watcher.Add(d.path); err != nil {
		return fmt.Errorf("watch device error:%v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				klog.Infof("fsnotify device event: %s %s", event.Name, event.Op.String())
				if event.Op == fsnotify.Create {
					dev := path.Base(event.Name)
					d.devices[dev] = &pluginapi.Device{
						ID:     dev,
						Health: pluginapi.Healthy,
					}
					d.notify <- struct{}{}
					klog.Infof("find new device [%s]", dev)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					dev := path.Base(event.Name)
					delete(d.devices, dev)
					d.notify <- struct{}{}
					klog.Infof("device [%s] removed", dev)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Errorf("fsnotify watch device failed:%v", err)
			}
		}
	}()
	return nil
}

func (d *DeviceMonitor) getDevices() []*pluginapi.Device {
	devices := make([]*pluginapi.Device, 0, len(d.devices))
	for _, dev := range d.devices {
		devices = append(devices, dev)
	}
	return devices
}

func String(devices []*pluginapi.Device) string {
	ids := make([]string, 0, len(devices))
	for _, dev := range devices {
		ids = append(ids, dev.ID)
	}
	return strings.Join(ids, ",")
}
