package main

import (
	"flag"
	"github.com/fsnotify/fsnotify"
	"github/forrest-tao/device-plugin-demo/pkg/utils"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()

	klog.Info("main func begins")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Fatalf("创建 watcher 失败: %v", err)
	}
	defer watcher.Close()
	stop := make(chan struct{})
	if err = utils.WatchKubelet(watcher, stop); err != nil {
		klog.Fatalf("utils.WatchKubelet failed :%v", err)
	}
	<-stop
	klog.Info("main func ends")
}
