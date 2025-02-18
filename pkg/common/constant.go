package common

import "time"

const (
	ResourceName      = "demo.com/xpu"
	DevicePath        = "/etc/xpu"
	DeviceSocket      = "xpu.sock"
	ConnectionTimeout = time.Second * 5
)
