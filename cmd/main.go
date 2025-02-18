package main

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"github/forrest-tao/device-plugin-demo/pkg/device_plugin"
	"github/forrest-tao/device-plugin-demo/pkg/utils"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()

	klog.Info("main func begins")
	//1.
	dp := device_plugin.NewXpuDevicePlugin()
	xputWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Fatalf("创建 watcher 失败: %v", err)
	}
	defer xputWatcher.Close()
	go dp.Run(xputWatcher)
	//2. 告诉 kubelet，我们这里有一个新的device plugin
	if err := dp.Register(); err != nil {
		klog.Fatalf("register to kubelet failed: %v", err)
	}
	//3.
	kubeletWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Fatalf("创建 watcher 失败: %v", err)
	}
	defer kubeletWatcher.Close()
	stop := make(chan struct{})
	if err = utils.WatchKubelet(kubeletWatcher, stop); err != nil {
		klog.Fatalf("utils.WatchKubelet failed :%v", err)
	}
	<-stop

	klog.Info("main func ends")
}
