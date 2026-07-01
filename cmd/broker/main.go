package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

type Room struct {
	HostIP   string
	HostPort string      // Host's UDP-port
	Channel  chan string // Канал, через который мы передадим IP:Port гостя обратно хосту
}

var rooms sync.Map

func handleCreate(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	port := r.URL.Query().Get("port")
	if code == "" || port == "" {
		http.Error(w, "Missing code or port", http.StatusBadRequest)
		return
	}

	hostIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid remote address", http.StatusInternalServerError)
		return
	}

	room := &Room{
		HostIP:   hostIP,
		HostPort: port,
		Channel:  make(chan string),
	}

	rooms.Store(code, room)
	log.Printf("[+] Room %s was created by %s:%s. Waiting for someone...\n", code, hostIP, port)

	guestAddr := <-room.Channel

	fmt.Fprintf(w, "%s", guestAddr)

	rooms.Delete(code)
}

func handleJoin(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	port := r.URL.Query().Get("port")
	if code == "" || port == "" {
		http.Error(w, "Missing code or port", http.StatusBadRequest)
		return
	}

	val, ok := rooms.Load(code)
	if !ok {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	room := val.(*Room)

	guestIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Invalid remote address", http.StatusInternalServerError)
		return
	}
	guestUDPAddr := net.JoinHostPort(guestIP, port)

	room.Channel <- guestUDPAddr

	fmt.Fprintf(w, "%s", net.JoinHostPort(room.HostIP, room.HostPort))
	log.Printf("[+] Guest %s entered %s.\n", guestUDPAddr, code)
}

func main() {
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/join", handleJoin)

	fmt.Println("\033[32m[+] Tube-Broker successfully launched :8080...\033[0m")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
