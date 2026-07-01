//go:build windows
// +build windows

package tun

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/songgao/water"
)

// CreateInterface attempts to create a TUN interface.
// If the TAP-Windows driver is missing, it downloads and installs it automatically.
func CreateInterface() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	// Target the official TAP-Windows driver
	config.PlatformSpecificParams = water.PlatformSpecificParams{
		ComponentID:   "tap0901",
		InterfaceName: "TubeTUN",
	}

	// 1. Try to create the interface first
	ifce, err := water.New(config)
	if err == nil {
		return ifce, nil // Success, driver is already installed
	}

	// 2. If it fails, assume the driver is missing and try to install it
	fmt.Println("\033[33m[*] TAP-Windows driver not found. Starting auto-installation...\033[0m")

	installerURL := "https://github.com/b4kh4/tube/raw/refs/heads/main/assets/tap-windows-9.21.2.exe" // Link for TAP installer
	tempPath := filepath.Join(os.TempDir(), "tap-windows-installer.exe")

	// Download the installer (250 KB)
	fmt.Println("[*] Downloading official driver (250 KB)...")
	if err := downloadFile(tempPath, installerURL); err != nil {
		return nil, fmt.Errorf("failed to download driver: %w", err)
	}
	defer os.Remove(tempPath) // Clean up the installer after we are done

	// Execute the installer silently (/S flag)
	fmt.Println("[*] Installing driver silently... Please wait.")
	cmd := exec.Command(tempPath, "/S")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("driver installation failed (Did you run as Administrator?): %w", err)
	}

	fmt.Println("\033[32m[+] Driver installed successfully!\033[0m")

	// 3. Try to create the interface again
	return water.New(config)
}

// ConfigureIP assigns the virtual IP address to the Windows network adapter
func ConfigureIP(ifceName string, vpnIP string) error {
	// Equivalent to: netsh interface ip set address name="TubeTUN" static 10.8.0.X 255.255.255.0
	cmd := exec.Command("netsh", "interface", "ip", "set", "address",
		"name="+ifceName, "static", vpnIP, "255.255.255.0")
	return cmd.Run()
}

// downloadFile is a helper function to fetch a file via HTTP
func downloadFile(filepath string, url string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}
