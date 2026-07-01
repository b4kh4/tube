package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/b4kh4/tube/internal/crypto"
	"github.com/b4kh4/tube/internal/tun"
	"github.com/songgao/water"
)

const brokerURL = "https://tube-broker.onrender.com" // Сloud URL

// Global variables to manage active session resources
var activeTUN *water.Interface
var activeUDP *net.UDPConn

// ==========================================
// 1. ENTRY POINT & SIGNAL HANDLING
// ==========================================

func main() {
	setupSignalHandler()
	runConsole()
}

func setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\033[31m[!] System interrupt received. Shutting down...\033[0m")
		stopSession()
		fmt.Println("\033[32m[+] Shutdown complete. Goodbye!\033[0m")
		os.Exit(0)
	}()
}

// ==========================================
// 2. INTERACTIVE CONSOLE (TUI)
// ==========================================

func runConsole() {
	scanner := bufio.NewScanner(os.Stdin)
	printWelcomeScreen()

	for {
		fmt.Print("tube > ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Split(input, " ")
		command := parts[0]

		switch command {
		case "/create":
			handleCreate()
		case "/join":
			if len(parts) < 2 {
				showError("Usage: /join <ROOM-CODE>", nil)
				continue
			}
			handleJoin(parts[1])
		case "/stop":
			stopSession()
		case "/help":
			printWelcomeScreen()
		case "/exit":
			fmt.Println("\033[33m[*] Exiting Tube VPN...\033[0m")
			stopSession()
			return
		default:
			showError("Unknown command. Type /help to see available commands.", nil)
		}
	}

	if err := scanner.Err(); err != nil {
		showError("Console input error", err)
	}
}

func printWelcomeScreen() {
	fmt.Println("\033[36m┌────────────────────────────────────────────────────────┐\033[0m")
	fmt.Println("\033[36m│\033[32m                    TUBE VPN v1.0                       \033[36m│\033[0m")
	fmt.Println("\033[36m│\033[33m             Secure P2P Encrypted Tunnel                \033[36m│\033[0m")
	fmt.Println("\033[36m├────────────────────────────────────────────────────────┤\033[0m")
	fmt.Println("\033[36m│\033[0m  Available Commands:                                   \033[36m│\033[0m")
	fmt.Println("\033[36m│\033[0m  /create          - Create a secure lobby (Host)       \033[36m│\033[0m")
	fmt.Println("\033[36m│\033[0m  /join <code>     - Connect using 16-character code    \033[36m│\033[0m")
	fmt.Println("\033[36m│\033[0m  /stop            - Disconnect current session         \033[36m│\033[0m")
	fmt.Println("\033[36m│\033[0m  /help            - Show this list of commands         \033[36m│\033[0m")
	fmt.Println("\033[36m│\033[0m  /exit            - Terminate program and exit         \033[36m│\033[0m")
	fmt.Println("\033[36m└────────────────────────────────────────────────────────┘\033[0m")
	fmt.Println()
}

func showError(message string, err error) {
	if err != nil {
		fmt.Printf("\033[31m[!] %s: %v\033[0m\n", message, err)
	} else {
		fmt.Printf("\033[31m[!] %s\033[0m\n", message)
	}
}

// ==========================================
// 3. NETWORK ENGINE CONTROLLERS
// ==========================================

func handleCreate() {
	stopSession()

	code := crypto.GenerateRoomCode()
	crypto.SetPassword(code)
	fmt.Printf("\033[32m[+] Room created successfully!\033[0m\n")
	fmt.Printf("\033[33m[+] Share this code with your friend: %s\033[0m\n", code)
	fmt.Println("[*] Waiting for peer to connect...")

	go func() {
		friendAddr, err := requestBroker("create", code, "4000")
		if err != nil {
			showError("\nBroker connection failed", err)
			return
		}
		startHost(friendAddr)
	}()
}

func handleJoin(code string) {
	stopSession()

	crypto.SetPassword(code)

	fmt.Printf("[*] Connecting to room %s...\n", code)
	hostAddr, err := requestBroker("join", code, "5000")
	if err != nil {
		showError("Failed to join room", err)
		return
	}
	startGuest(hostAddr)
}

func startHost(friendAddr string) {
	fmt.Printf("\033[33m[*] Establishing P2P link as Host (IP: 10.8.0.1)...\033[0m\n")

	localAddr, err := net.ResolveUDPAddr("udp", ":4000")
	if err != nil {
		showError("Local address resolution failed", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		showError("Could not bind UDP port 4000", err)
		return
	}
	activeUDP = udpConn

	remoteAddr, err := net.ResolveUDPAddr("udp", friendAddr)
	if err != nil {
		showError("Peer address resolution failed", err)
		return
	}

	ifce, err := tun.Start(udpConn, remoteAddr, "10.8.0.1")
	if err != nil {
		showError("Tunnel initialization failed", err)
		return
	}
	activeTUN = ifce

	fmt.Println("\033[32m[+] Secure P2P Tunnel active!\033[0m")
}

func startGuest(friendAddr string) {
	fmt.Printf("\033[33m[*] Establishing P2P link as Guest (IP: 10.8.0.2)...\033[0m\n")

	localAddr, err := net.ResolveUDPAddr("udp", ":5000")
	if err != nil {
		showError("Local address resolution failed", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		showError("Could not bind UDP port 5000", err)
		return
	}
	activeUDP = udpConn

	remoteAddr, err := net.ResolveUDPAddr("udp", friendAddr)
	if err != nil {
		showError("Host address resolution failed", err)
		return
	}

	ifce, err := tun.Start(udpConn, remoteAddr, "10.8.0.2")
	if err != nil {
		showError("Tunnel initialization failed", err)
		return
	}
	activeTUN = ifce

	fmt.Println("\033[32m[+] Secure P2P Tunnel connected!\033[0m")
}

func stopSession() {
	if activeTUN == nil && activeUDP == nil {
		return
	}

	fmt.Println("\033[33m[*] Disconnecting active session...\033[0m")
	if activeTUN != nil {
		activeTUN.Close()
		activeTUN = nil
	}
	if activeUDP != nil {
		activeUDP.Close()
		activeUDP = nil
	}
	fmt.Println("\033[32m[+] Session terminated.\033[0m")
}

// ==========================================
// 4. BROKER API CLIENT
// ==========================================

func requestBroker(endpoint string, code string, port string) (string, error) {
	url := fmt.Sprintf("%s/%s?code=%s&port=%s", brokerURL, endpoint, code, port)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("broker returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
