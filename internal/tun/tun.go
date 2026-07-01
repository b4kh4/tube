package tun

import (
	"fmt"
	"log"
	"net"

	"github.com/b4kh4/tube/internal/crypto" // Make sure this path matches your go.mod!
	"github.com/songgao/water"
)

// Start initializes the virtual interface and begins background routing
func Start(udpConn *net.UDPConn, remoteAddr *net.UDPAddr, vpnIP string) (*water.Interface, error) {
	fmt.Println("[*] Initializing virtual network adapter...")

	ifce, err := CreateInterface()
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN interface: %w", err)
	}

	// Make sure ConfigureIP exists in your tun_linux.go and tun_windows.go!
	err = ConfigureIP(ifce.Name(), vpnIP)
	if err != nil {
		ifce.Close() // Clean up if OS configuration fails
		return nil, fmt.Errorf("OS routing configuration failed: %w", err)
	}
	fmt.Printf("[+] Adapter '%s' provisioned with IP %s\n", ifce.Name(), vpnIP)

	// Launch bi-directional traffic handlers in background
	go udpToTun(ifce, udpConn)
	go tunToUDP(ifce, udpConn, remoteAddr)

	return ifce, nil
}

// tunToUDP captures local OS traffic, encrypts it, and sends it over UDP
func tunToUDP(ifce *water.Interface, udpConn *net.UDPConn, remoteAddr *net.UDPAddr) {
	packet := make([]byte, 2000)

	for {
		n, err := ifce.Read(packet)
		if err != nil {
			// If the interface is closed (e.g. by /stop), exit the goroutine cleanly
			return
		}

		// IPv4 filtering (first byte shifted right by 4 must equal 4)
		if packet[0]>>4 != 4 {
			continue
		}

		encrypted, err := crypto.Encrypt(packet[:n])
		if err != nil {
			log.Println("[!] Encryption failed:", err)
			continue
		}

		_, err = udpConn.WriteToUDP(encrypted, remoteAddr)
		if err != nil {
			// Silent continue: normal behavior if peer is temporarily unreachable
			continue
		}
	}
}

// udpToTun receives encrypted UDP traffic, decrypts it, and injects it into the OS
func udpToTun(ifce *water.Interface, udpConn *net.UDPConn) {
	packet := make([]byte, 2000)

	for {
		n, _, err := udpConn.ReadFromUDP(packet)
		if err != nil {
			// If the UDP socket is closed (e.g. by /stop), exit the goroutine cleanly
			return
		}

		decrypted, err := crypto.Decrypt(packet[:n])
		if err != nil {
			// Drop packet: likely unauthorized or corrupted data
			continue
		}

		_, err = ifce.Write(decrypted)
		if err != nil {
			log.Println("[!] TUN injection failed:", err)
		}
	}
}
