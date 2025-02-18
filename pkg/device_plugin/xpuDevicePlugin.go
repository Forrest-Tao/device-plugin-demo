package device_plugin

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github/forrest-tao/device-plugin-demo/pkg/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"log"
	"net"
	"os"
	"path"
	"syscall"
	"time"
)

type XpuDevicePlugin struct {
	server *grpc.Server
	stop   chan struct{} // this channel signals to stop the device plugin
	dm     *DeviceMonitor
}

func NewXpuDevicePlugin() *XpuDevicePlugin {
	return &XpuDevicePlugin{
		server: grpc.NewServer(grpc.EmptyServerOption{}),
		stop:   make(chan struct{}),
		dm:     NewDeviceMonitor(common.DevicePath),
	}
}

// dial establishes the gRPC communication with the registered device plugin.
func connect(socketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c, err := grpc.DialContext(ctx, socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			if deadline, ok := ctx.Deadline(); ok {
				return net.DialTimeout("unix", addr, time.Until(deadline))
			}
			return net.DialTimeout("unix", addr, common.ConnectionTimeout)
		}),
	)
	if err != nil {
		klog.Errorf("connect exceed,err: %v", err)
		return nil, err
	}

	return c, nil
}

func (x *XpuDevicePlugin) Run(wather *fsnotify.Watcher) {
	if err := x.dm.List(); err != nil {
		log.Fatalf("list device failed %v", err)
	}

	if err := x.dm.Watch(wather); err != nil {
		log.Fatal("watch devices error", err)
	}
	pluginapi.RegisterDevicePluginServer(x.server, x)
	socket := path.Join(pluginapi.DevicePluginPath, common.DeviceSocket)
	if err := syscall.Unlink(socket); err != nil && !os.IsNotExist(err) {
		log.Fatal(errors.WithMessagef(err, "delete socket %s failed", socket))
	}

	sock, err := net.Listen("unix", socket)
	if err != nil {
		log.Fatal(errors.WithMessagef(err, "listen unix %s failed", socket))
	}

	go x.server.Serve(sock)

	// Wait for server to start by launching a blocking connection
	conn, err := connect(socket, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	conn.Close()
}
