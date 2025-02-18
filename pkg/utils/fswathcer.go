package utils

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"path"
	"path/filepath"
	"time"
)

// WatchKubelet restart device plugin when kubelet restarted
func WatchKubelet(watcher *fsnotify.Watcher, stop chan<- struct{}) error {
	kubeletDir := pluginapi.DevicePluginPath
	if err := listFiles(kubeletDir); err != nil {
		return errors.WithMessagef(err, "unable to listfiles")
	}
	if err := watcher.Add(kubeletDir); err != nil {
		return errors.WithMessagef(err, "Unable to add path %s to watcher", kubeletDir)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				printEvent(event)
				if path.Base(event.Name) == "kubelet.sock" && event.Op == fsnotify.Create {
					klog.Warning("inotify: kubelet.sock created, restarting.")
					stop <- struct{}{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Errorf("fsnotify failed restarting,detail:%v", err)
			}
		}
	}()
	return nil
}

func listFiles(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		fmt.Printf("文件名: %-20s 大小: %-10d 修改时间: %s\n",
			file.Name(),
			info.Size(),
			info.ModTime().Format("2006-01-02 15:04:05"))
	}
	return nil
}

func printEvent(event fsnotify.Event) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	fileInfo := "文件已删除"
	if event.Op != fsnotify.Remove {
		if info, err := os.Stat(event.Name); err == nil {
			fileInfo = fmt.Sprintf("大小: %d bytes", info.Size())
		}
	}

	op := getOperationType(event.Op)

	klog.Info("[%s] 文件: %-30s 操作: %-10s %s\n",
		timestamp,
		filepath.Base(event.Name),
		op,
		fileInfo)
}

func getOperationType(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "创建"
	case op&fsnotify.Write == fsnotify.Write:
		return "修改"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "删除"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "重命名"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "权限修改"
	default:
		return "未知操作"
	}
}
