//go:build linux
// +build linux

package tun

import (
	"os/exec"

	"github.com/songgao/water"
)

// CreateInterface creates a virtual TUN interface on Linux
func CreateInterface() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	return water.New(config)
}

// ConfigureIP sets up the IP address and brings the interface UP on Linux
func ConfigureIP(ifceName string, vpnIP string) error {
	// Equivalent to: ip addr add 10.8.0.X/24 dev tun0
	cmdAddr := exec.Command("ip", "addr", "add", vpnIP+"/24", "dev", ifceName)
	if err := cmdAddr.Run(); err != nil {
		return err
	}

	// Equivalent to: ip link set dev tun0 up
	cmdUp := exec.Command("ip", "link", "set", "dev", ifceName, "up")
	return cmdUp.Run()
}
